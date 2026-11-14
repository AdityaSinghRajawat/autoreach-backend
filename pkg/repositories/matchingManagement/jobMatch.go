package matchingManagement

import (
	"context"
	"fmt"

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
}

func NewJobMatchRepo(postgres *db.PostgresConn) (*JobMatchRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new job match repo: %w", config.ErrMissingDependency)
	}

	return &JobMatchRepo{
		postgres:         postgres,
		jobMatchesTable:  config.GetTableName(config.TableJobMatches),
		jobPostingsTable: config.GetTableName(config.TableJobPostings),
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

func (r *JobMatchRepo) ListForCandidate(
	ctx context.Context,
	candidateID string,
	minScore float64,
	page sharedTypes.PageParams,
) ([]matchingTypes.MatchedJob, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT job.id, job.company_name, job.job_title, job.location,
			match.match_score::float8, job.ats_type, 0, NULL::text, job.created_at
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
