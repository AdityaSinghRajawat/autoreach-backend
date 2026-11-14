package matchingManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
)

type JobMatchRepo struct {
	postgres         *db.PostgresConn
	jobMatchesTable  string
	jobPostingsTable string
	jobContactsTable string
}

func NewJobMatchRepo(postgres *db.PostgresConn) (*JobMatchRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new job match repo: %w", config.ErrMissingDependency)
	}

	return &JobMatchRepo{
		postgres:         postgres,
		jobMatchesTable:  config.GetTableName(config.TableJobMatches),
		jobPostingsTable: config.GetTableName(config.TableJobPostings),
		jobContactsTable: config.GetTableName(config.TableJobContacts),
	}, nil
}

func (r *JobMatchRepo) UpsertMany(ctx context.Context, matches []*matchingModel.JobMatch) error {
	if len(matches) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, match := range matches {
		batch.Queue(`
			INSERT INTO `+r.jobMatchesTable+` (
				id, candidate_id, job_id, resume_id, raw_similarity, match_score,
				created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (candidate_id, job_id) DO UPDATE SET
				resume_id = EXCLUDED.resume_id,
				raw_similarity = EXCLUDED.raw_similarity,
				match_score = EXCLUDED.match_score,
				updated_at = EXCLUDED.updated_at
		`,
			match.ID, match.CandidateID, match.JobID, match.ResumeID,
			match.RawSimilarity, match.MatchScore, match.CreatedAt, match.UpdatedAt,
		)
	}

	if err := r.postgres.Executor(ctx).SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("upsert job matches: %w", err)
	}

	return nil
}

func (r *JobMatchRepo) DeleteStaleForCandidate(
	ctx context.Context,
	candidateID, currentResumeID string,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		DELETE FROM `+r.jobMatchesTable+`
		WHERE candidate_id = $1 AND resume_id <> $2 AND fanned_out_at IS NULL
	`, candidateID, currentResumeID)
	if err != nil {
		return fmt.Errorf("delete stale job matches: %w", err)
	}

	return nil
}

func (r *JobMatchRepo) Exists(ctx context.Context, candidateID, jobID string) (bool, error) {
	var exists bool

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM `+r.jobMatchesTable+` WHERE candidate_id = $1 AND job_id = $2
		)
	`, candidateID, jobID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check job match: %w", err)
	}

	return exists, nil
}

func (r *JobMatchRepo) ClaimFanOut(
	ctx context.Context,
	candidateID string,
	minScore float64,
	claimedAt time.Time,
) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		UPDATE `+r.jobMatchesTable+`
		SET fanned_out_at = $3, updated_at = $3
		WHERE candidate_id = $1 AND match_score >= $2 AND fanned_out_at IS NULL
		RETURNING job_id::text
	`, candidateID, minScore, claimedAt)
	if err != nil {
		return nil, fmt.Errorf("claim job match fan-out: %w", err)
	}

	jobIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("claim job match fan-out: %w", err)
	}

	return jobIDs, nil
}

func (r *JobMatchRepo) ReleaseFanOut(ctx context.Context, candidateID, jobID string) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.jobMatchesTable+`
		SET fanned_out_at = NULL
		WHERE candidate_id = $1 AND job_id = $2
	`, candidateID, jobID)
	if err != nil {
		return fmt.Errorf("release job match fan-out: %w", err)
	}

	return nil
}

func (r *JobMatchRepo) ListForCandidate(
	ctx context.Context,
	candidateID string,
	minScore float64,
	page sharedTypes.PageParams,
) ([]matchingTypes.MatchedJob, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT job.id, job.company_name, job.job_title, job.location,
			match.match_score::float8, job.ats_type,
			(SELECT COUNT(*) FROM `+r.jobContactsTable+` AS link WHERE link.job_id = job.id)::int,
			NULL::text, job.created_at
		FROM `+r.jobMatchesTable+` AS match
		JOIN `+r.jobPostingsTable+` AS job ON job.id = match.job_id
		WHERE match.candidate_id = $1 AND match.match_score >= $2
		ORDER BY match.match_score DESC, job.created_at DESC, job.id
		LIMIT $3 OFFSET $4
	`, candidateID, minScore, page.PerPage, page.Offset())
	if err != nil {
		return nil, fmt.Errorf("list matched jobs: %w", err)
	}

	jobs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[matchingTypes.MatchedJob])
	if err != nil {
		return nil, fmt.Errorf("list matched jobs: %w", err)
	}

	return jobs, nil
}

func (r *JobMatchRepo) CountForCandidate(
	ctx context.Context,
	candidateID string,
	minScore float64,
) (int, error) {
	var total int

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM `+r.jobMatchesTable+`
		WHERE candidate_id = $1 AND match_score >= $2
	`, candidateID, minScore).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count matched jobs: %w", err)
	}

	return total, nil
}
