package applicationManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
)

const applicationColumns = `id, candidate_id, job_id, resume_id, match_score, status, ats_type, apply_url,
	prepared_answers, missing_answers, execution_logs, external_confirmation, attempt_count,
	failure_reason, submitted_at, created_at, updated_at`

type ApplicationRepo struct {
	postgres          *db.PostgresConn
	applicationsTable string
	jobMatchesTable   string
	jobPostingsTable  string
	resumesTable      string
}

func NewApplicationRepo(postgres *db.PostgresConn) (*ApplicationRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new application repo: %w", config.ErrMissingDependency)
	}

	return &ApplicationRepo{
		postgres:          postgres,
		applicationsTable: config.GetTableName(config.TableApplications),
		jobMatchesTable:   config.GetTableName(config.TableJobMatches),
		jobPostingsTable:  config.GetTableName(config.TableJobPostings),
		resumesTable:      config.GetTableName(config.TableCandidateResumes),
	}, nil
}

func (r *ApplicationRepo) Upsert(
	ctx context.Context,
	application *applicationModel.Application,
) (*applicationModel.Application, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.applicationsTable+` (`+applicationColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (id) DO UPDATE SET
			resume_id = EXCLUDED.resume_id,
			status = EXCLUDED.status,
			prepared_answers = EXCLUDED.prepared_answers,
			missing_answers = EXCLUDED.missing_answers,
			execution_logs = EXCLUDED.execution_logs,
			external_confirmation = EXCLUDED.external_confirmation,
			attempt_count = EXCLUDED.attempt_count,
			failure_reason = EXCLUDED.failure_reason,
			submitted_at = EXCLUDED.submitted_at,
			updated_at = EXCLUDED.updated_at
	`,
		application.ID, application.CandidateID, application.JobID, application.ResumeID,
		application.MatchScore, string(application.Status), application.ATSType, application.ApplyURL,
		string(application.PreparedAnswers), string(application.MissingAnswers),
		string(
			application.ExecutionLogs,
		), application.ExternalConfirmation, application.AttemptCount,
		application.FailureReason, application.SubmittedAt, application.CreatedAt, application.UpdatedAt,
	)
	if db.IsUniqueViolation(err) {
		return nil, fmt.Errorf("upsert application: %w", config.ErrDuplicateRecord)
	}

	if err != nil {
		return nil, fmt.Errorf("upsert application: %w", err)
	}

	return application, nil
}

func (r *ApplicationRepo) QueueFromMatches(
	ctx context.Context,
	candidateID string,
	minScore float64,
	now time.Time,
) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		INSERT INTO `+r.applicationsTable+` (candidate_id, job_id, resume_id, match_score, status, ats_type,
			apply_url, created_at, updated_at)
		SELECT match.candidate_id, match.job_id, match.resume_id, match.match_score, 'QUEUED',
			job.ats_type, job.external_url, $3, $3
		FROM `+r.jobMatchesTable+` AS match
		JOIN `+r.jobPostingsTable+` AS job ON job.id = match.job_id AND job.status = 'ACTIVE'
		WHERE match.candidate_id = $1 AND match.match_score >= $2
		ON CONFLICT ON CONSTRAINT unique_application_candidate_job DO NOTHING
		RETURNING id
	`, candidateID, minScore, now)
	if err != nil {
		return nil, fmt.Errorf("queue applications: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("queue applications: %w", err)
	}

	return ids, nil
}

func (r *ApplicationRepo) MatchScore(
	ctx context.Context,
	candidateID, jobID string,
) (float64, error) {
	var score float64

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COALESCE((SELECT match_score::float8 FROM `+r.jobMatchesTable+`
			WHERE candidate_id = $1 AND job_id = $2), 0)
	`, candidateID, jobID).Scan(&score)
	if err != nil {
		return 0, fmt.Errorf("find match score: %w", err)
	}

	return score, nil
}

func (r *ApplicationRepo) FindByID(
	ctx context.Context,
	applicationID string,
) (*applicationModel.Application, error) {
	return r.findOne(ctx, `id = $1`, applicationID)
}

func (r *ApplicationRepo) FindByCandidate(
	ctx context.Context,
	candidateID, applicationID string,
) (*applicationModel.Application, error) {
	return r.findOne(ctx, `candidate_id = $1 AND id = $2`, candidateID, applicationID)
}

func (r *ApplicationRepo) ListByCandidate(
	ctx context.Context,
	candidateID, status string,
	page sharedTypes.PageParams,
) ([]*applicationModel.Application, int, error) {
	condition := `candidate_id = $1 AND ($2 = '' OR status::text = $2)`

	var total int

	if err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM `+r.applicationsTable+` WHERE `+condition, candidateID, status,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}

	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+applicationColumns+` FROM `+r.applicationsTable+` WHERE `+condition+`
		ORDER BY created_at DESC, id LIMIT $3 OFFSET $4
	`, candidateID, status, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list applications: %w", err)
	}

	applications, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*applicationModel.Application, error) {
			return scanApplication(row)
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list applications: %w", err)
	}

	return applications, total, nil
}

// Claim gives one worker an application that is queued, or whose previous run stalled before
// staleBefore; it reports false when another run holds it or it is no longer pending.
func (r *ApplicationRepo) Claim(
	ctx context.Context,
	applicationID string,
	staleBefore, now time.Time,
) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.applicationsTable+`
		SET status = 'IN_PROGRESS', attempt_count = attempt_count + 1, updated_at = $3
		WHERE id = $1 AND (status = 'QUEUED' OR (status = 'IN_PROGRESS' AND updated_at < $2))
	`, applicationID, staleBefore, now)
	if err != nil {
		return false, fmt.Errorf("claim application: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

// Release returns a claimed application to the queue so a retry can claim it again.
func (r *ApplicationRepo) Release(ctx context.Context, applicationID string, now time.Time) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.applicationsTable+` SET status = 'QUEUED', updated_at = $2
		WHERE id = $1 AND status = 'IN_PROGRESS'
	`, applicationID, now)
	if err != nil {
		return fmt.Errorf("release application: %w", err)
	}

	return nil
}

func (r *ApplicationRepo) ListStalePendingIDs(
	ctx context.Context,
	staleBefore time.Time,
	limit int,
) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT id FROM `+r.applicationsTable+`
		WHERE status IN ('QUEUED', 'IN_PROGRESS') AND updated_at < $1
		ORDER BY updated_at LIMIT $2
	`, staleBefore, limit)
	if err != nil {
		return nil, fmt.Errorf("list stale applications: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list stale applications: %w", err)
	}

	return ids, nil
}

func (r *ApplicationRepo) findOne(
	ctx context.Context,
	condition string,
	arguments ...any,
) (*applicationModel.Application, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+applicationColumns+` FROM `+r.applicationsTable+` WHERE `+condition, arguments...)

	return db.FindOne(row, scanApplication, "find application")
}

func scanApplication(row db.RowScanner) (*applicationModel.Application, error) {
	application := applicationModel.Application{}

	var preparedAnswers, missingAnswers, executionLogs []byte

	err := row.Scan(
		&application.ID, &application.CandidateID, &application.JobID, &application.ResumeID,
		&application.MatchScore, &application.Status, &application.ATSType, &application.ApplyURL,
		&preparedAnswers, &missingAnswers, &executionLogs, &application.ExternalConfirmation,
		&application.AttemptCount, &application.FailureReason, &application.SubmittedAt,
		&application.CreatedAt, &application.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	application.PreparedAnswers = preparedAnswers
	application.MissingAnswers = missingAnswers
	application.ExecutionLogs = executionLogs

	return &application, nil
}
