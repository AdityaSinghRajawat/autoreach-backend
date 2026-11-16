package jobManagement

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

const jobPostingColumns = `id, job_hash, source::text, company_name, company_domain, job_title,
	normalized_title, job_description, description_hash, location, department,
	salary_min::float8, salary_max::float8, salary_currency, ats_type, external_url,
	external_job_id, status::text, (embedding IS NOT NULL) AS is_embedded,
	posted_at, last_seen_at, created_at, updated_at, job_source_id`

type JobPostingRepo struct {
	postgres         *db.PostgresConn
	jobPostingsTable string
}

func NewJobPostingRepo(postgres *db.PostgresConn) (*JobPostingRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new job posting repo: %w", config.ErrMissingDependency)
	}

	return &JobPostingRepo{
		postgres:         postgres,
		jobPostingsTable: config.GetTableName(config.TableJobPostings),
	}, nil
}

func (r *JobPostingRepo) UpsertMany(
	ctx context.Context,
	jobs []*jobModel.JobPosting,
) ([]jobTypes.UpsertedJob, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	statement := `
		INSERT INTO ` + r.jobPostingsTable + ` AS existing (
			job_hash, source, company_name, company_domain, job_title, normalized_title,
			job_description, description_hash, location, department, salary_min, salary_max,
			salary_currency, ats_type, external_url, external_job_id, raw_payload, posted_at,
			last_seen_at, created_at, updated_at, job_source_id
		)
		VALUES (
			$1, $2::job_source, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17::jsonb, $18, $19, $20, $21, $22
		)
		ON CONFLICT (job_hash) DO UPDATE SET
			job_description = EXCLUDED.job_description,
			embedding = CASE
				WHEN existing.description_hash = EXCLUDED.description_hash THEN existing.embedding
			END,
			embedded_at = CASE
				WHEN existing.description_hash = EXCLUDED.description_hash THEN existing.embedded_at
			END,
			description_hash = EXCLUDED.description_hash,
			location = COALESCE(EXCLUDED.location, existing.location),
			department = COALESCE(EXCLUDED.department, existing.department),
			salary_min = COALESCE(EXCLUDED.salary_min, existing.salary_min),
			salary_max = COALESCE(EXCLUDED.salary_max, existing.salary_max),
			salary_currency = COALESCE(EXCLUDED.salary_currency, existing.salary_currency),
			posted_at = COALESCE(existing.posted_at, EXCLUDED.posted_at),
			status = ` + db.Enum(modelHelpers.JobStatusActive) + `,
			job_source_id = COALESCE(EXCLUDED.job_source_id, existing.job_source_id),
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, (embedding IS NULL) AS needs_embedding`

	batch := &pgx.Batch{}
	for _, job := range jobs {
		batch.Queue(
			statement,
			job.JobHash, job.Source, job.CompanyName, job.CompanyDomain, job.JobTitle,
			job.NormalizedTitle, job.JobDescription, job.DescriptionHash, job.Location,
			job.Department, job.SalaryMin, job.SalaryMax, job.SalaryCurrency, job.ATSType,
			job.ExternalURL, job.ExternalJobID, string(job.RawPayload), job.PostedAt,
			job.LastSeenAt, job.CreatedAt, job.UpdatedAt, job.JobSourceID,
		)
	}

	results := r.postgres.Executor(ctx).SendBatch(ctx, batch)

	upserted := make([]jobTypes.UpsertedJob, 0, len(jobs))
	for range jobs {
		row := jobTypes.UpsertedJob{}
		if err := results.QueryRow().Scan(&row.ID, &row.NeedsEmbedding); err != nil {
			_ = results.Close()

			return nil, fmt.Errorf("upsert job postings: %w", err)
		}

		upserted = append(upserted, row)
	}

	if err := results.Close(); err != nil {
		return nil, fmt.Errorf("upsert job postings: %w", err)
	}

	return upserted, nil
}

func (r *JobPostingRepo) FindByID(ctx context.Context, jobID string) (*jobModel.JobPosting, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+jobPostingColumns+` FROM `+r.jobPostingsTable+` WHERE id = $1 LIMIT 1
	`, jobID)

	return db.FindOne(row, scanJobPosting, "find job posting")
}

func (r *JobPostingRepo) ListByIDs(
	ctx context.Context,
	jobIDs []string,
) ([]*jobModel.JobPosting, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+jobPostingColumns+` FROM `+r.jobPostingsTable+` WHERE id = ANY($1)
	`, jobIDs)
	if err != nil {
		return nil, fmt.Errorf("list jobs by id: %w", err)
	}

	jobs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*jobModel.JobPosting, error) {
		return scanJobPosting(row)
	})
	if err != nil {
		return nil, fmt.Errorf("list jobs by id: %w", err)
	}

	return jobs, nil
}

func (r *JobPostingRepo) ListEmbeddingSources(
	ctx context.Context,
	jobIDs []string,
) ([]jobTypes.JobEmbeddingSource, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT id, company_name, job_title, location, department, job_description, description_hash
		FROM `+r.jobPostingsTable+`
		WHERE id = ANY($1::uuid[]) AND embedding IS NULL AND status = `+sqlJobStatusActive+`
		ORDER BY id
	`, jobIDs)
	if err != nil {
		return nil, fmt.Errorf("list job embedding sources: %w", err)
	}

	sources, err := pgx.CollectRows(rows, scanJobEmbeddingSource)
	if err != nil {
		return nil, fmt.Errorf("list job embedding sources: %w", err)
	}

	return sources, nil
}

func (r *JobPostingRepo) ListStalePendingEmbeddingIDs(
	ctx context.Context,
	staleBefore time.Time,
	limit int,
) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT id::text FROM `+r.jobPostingsTable+`
		WHERE embedding IS NULL AND status = `+sqlJobStatusActive+` AND updated_at < $1
		ORDER BY updated_at
		LIMIT $2
	`, staleBefore, limit)
	if err != nil {
		return nil, fmt.Errorf("list stale pending embeddings: %w", err)
	}

	jobIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list stale pending embeddings: %w", err)
	}

	return jobIDs, nil
}

func (r *JobPostingRepo) UpdateEmbeddings(
	ctx context.Context,
	updates []jobTypes.JobEmbeddingUpdate,
	model string,
	embeddedAt time.Time,
) (int, error) {
	if len(updates) == 0 {
		return 0, nil
	}

	batch := &pgx.Batch{}
	for _, update := range updates {
		batch.Queue(`
			UPDATE `+r.jobPostingsTable+`
			SET embedding = $2::vector, embedding_model = $3, embedded_at = $4, updated_at = $4
			WHERE id = $1 AND description_hash = $5
		`, update.ID, db.FormatVector(update.Embedding), model, embeddedAt, update.DescriptionHash)
	}

	results := r.postgres.Executor(ctx).SendBatch(ctx, batch)

	updated := 0

	for range updates {
		tag, err := results.Exec()
		if err != nil {
			_ = results.Close()

			return updated, fmt.Errorf("update job embeddings: %w", err)
		}

		updated += int(tag.RowsAffected())
	}

	if err := results.Close(); err != nil {
		return updated, fmt.Errorf("update job embeddings: %w", err)
	}

	return updated, nil
}

func (r *JobPostingRepo) ListNearestActive(
	ctx context.Context,
	vector string,
	limit int,
	searchBreadth int,
) ([]matchingTypes.JobSimilarity, error) {
	executor := r.postgres.Executor(ctx)

	if _, err := executor.Exec(
		ctx,
		`SET LOCAL hnsw.ef_search = `+strconv.Itoa(searchBreadth),
	); err != nil {
		return nil, fmt.Errorf("set vector search breadth: %w", err)
	}

	rows, err := executor.Query(ctx, `
		SELECT id, (1 - (embedding <=> $1::vector))::float8
		FROM `+r.jobPostingsTable+`
		WHERE status = `+sqlJobStatusActive+` AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vector, limit)
	if err != nil {
		return nil, fmt.Errorf("list nearest active jobs: %w", err)
	}

	similarities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[matchingTypes.JobSimilarity])
	if err != nil {
		return nil, fmt.Errorf("list nearest active jobs: %w", err)
	}

	return similarities, nil
}

func scanJobPosting(row db.RowScanner) (*jobModel.JobPosting, error) {
	job := jobModel.JobPosting{}

	err := row.Scan(
		&job.ID, &job.JobHash, &job.Source, &job.CompanyName, &job.CompanyDomain, &job.JobTitle,
		&job.NormalizedTitle, &job.JobDescription, &job.DescriptionHash, &job.Location,
		&job.Department, &job.SalaryMin, &job.SalaryMax, &job.SalaryCurrency, &job.ATSType,
		&job.ExternalURL, &job.ExternalJobID, &job.Status, &job.IsEmbedded,
		&job.PostedAt, &job.LastSeenAt, &job.CreatedAt, &job.UpdatedAt, &job.JobSourceID,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func scanJobEmbeddingSource(row pgx.CollectableRow) (jobTypes.JobEmbeddingSource, error) {
	source := jobTypes.JobEmbeddingSource{}

	err := row.Scan(
		&source.ID, &source.CompanyName, &source.JobTitle, &source.Location,
		&source.Department, &source.JobDescription, &source.DescriptionHash,
	)

	return source, err
}
