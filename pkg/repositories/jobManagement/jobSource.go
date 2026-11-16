package jobManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

const jobSourceColumns = `id, board_type, board_token, search_query, search_location, region, company_name,
	company_domain, is_active, last_polled_at, last_job_count, last_error, created_at, updated_at`

type JobSourceRepo struct {
	postgres         *db.PostgresConn
	jobSourcesTable  string
	jobPostingsTable string
}

func NewJobSourceRepo(postgres *db.PostgresConn) (*JobSourceRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new job source repo: %w", config.ErrMissingDependency)
	}

	return &JobSourceRepo{
		postgres:         postgres,
		jobSourcesTable:  config.GetTableName(config.TableJobSources),
		jobPostingsTable: config.GetTableName(config.TableJobPostings),
	}, nil
}

func (r *JobSourceRepo) Upsert(
	ctx context.Context,
	source *jobModel.JobSource,
) (*jobModel.JobSource, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.jobSourcesTable+` (`+jobSourceColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			search_query = EXCLUDED.search_query,
			search_location = EXCLUDED.search_location,
			company_name = EXCLUDED.company_name,
			company_domain = EXCLUDED.company_domain,
			is_active = EXCLUDED.is_active,
			last_polled_at = EXCLUDED.last_polled_at,
			last_job_count = EXCLUDED.last_job_count,
			last_error = EXCLUDED.last_error,
			updated_at = EXCLUDED.updated_at
	`,
		source.ID, string(source.BoardType), source.BoardToken, source.SearchQuery, source.SearchLocation,
		string(
			source.Region,
		), source.CompanyName, source.CompanyDomain, source.IsActive, source.LastPolledAt,
		source.LastJobCount, source.LastError, source.CreatedAt, source.UpdatedAt,
	)
	if db.IsUniqueViolation(err) {
		return nil, fmt.Errorf("upsert job source: %w", config.ErrDuplicateRecord)
	}

	if err != nil {
		return nil, fmt.Errorf("upsert job source: %w", err)
	}

	return source, nil
}

func (r *JobSourceRepo) FindByID(
	ctx context.Context,
	sourceID string,
) (*jobModel.JobSource, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+jobSourceColumns+` FROM `+r.jobSourcesTable+` WHERE id = $1
	`, sourceID)

	return db.FindOne(row, scanJobSource, "find job source")
}

func (r *JobSourceRepo) List(ctx context.Context) ([]*jobModel.JobSource, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+jobSourceColumns+` FROM `+r.jobSourcesTable+`
		ORDER BY board_type, board_token, search_query NULLS FIRST
	`)
	if err != nil {
		return nil, fmt.Errorf("list job sources: %w", err)
	}

	sources, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*jobModel.JobSource, error) {
			return scanJobSource(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list job sources: %w", err)
	}

	return sources, nil
}

func (r *JobSourceRepo) ListActiveIDs(ctx context.Context) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT id FROM `+r.jobSourcesTable+` WHERE is_active ORDER BY last_polled_at NULLS FIRST
	`)
	if err != nil {
		return nil, fmt.Errorf("list active job sources: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list active job sources: %w", err)
	}

	return ids, nil
}

func (r *JobSourceRepo) Delete(ctx context.Context, sourceID string) (bool, error) {
	tag, err := r.postgres.Executor(ctx).
		Exec(ctx, `DELETE FROM `+r.jobSourcesTable+` WHERE id = $1`, sourceID)
	if err != nil {
		return false, fmt.Errorf("delete job source: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *JobSourceRepo) ListActiveExternalIDs(
	ctx context.Context,
	sourceID string,
) (map[string]struct{}, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT external_job_id FROM `+r.jobPostingsTable+`
		WHERE job_source_id = $1 AND status = 'ACTIVE' AND external_job_id IS NOT NULL
	`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("list known board jobs: %w", err)
	}

	externalIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list known board jobs: %w", err)
	}

	known := make(map[string]struct{}, len(externalIDs))
	for _, externalID := range externalIDs {
		known[externalID] = struct{}{}
	}

	return known, nil
}

func (r *JobSourceRepo) MarkSeen(
	ctx context.Context,
	sourceID string,
	externalIDs []string,
	now time.Time,
) (int, error) {
	if len(externalIDs) == 0 {
		return 0, nil
	}

	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.jobPostingsTable+` SET last_seen_at = $3
		WHERE job_source_id = $1 AND status = 'ACTIVE' AND external_job_id = ANY($2)
	`, sourceID, externalIDs, now)
	if err != nil {
		return 0, fmt.Errorf("mark board jobs seen: %w", err)
	}

	return int(tag.RowsAffected()), nil
}

// CloseUnseen closes the source's active jobs not seen since closeBefore.
func (r *JobSourceRepo) CloseUnseen(
	ctx context.Context,
	sourceID string,
	closeBefore, now time.Time,
) (int, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.jobPostingsTable+` SET status = 'CLOSED', updated_at = $3
		WHERE job_source_id = $1 AND status = 'ACTIVE' AND last_seen_at < $2
	`, sourceID, closeBefore, now)
	if err != nil {
		return 0, fmt.Errorf("close unseen jobs: %w", err)
	}

	return int(tag.RowsAffected()), nil
}

func scanJobSource(row db.RowScanner) (*jobModel.JobSource, error) {
	source := jobModel.JobSource{}

	err := row.Scan(
		&source.ID,
		&source.BoardType,
		&source.BoardToken,
		&source.SearchQuery,
		&source.SearchLocation,
		&source.Region,
		&source.CompanyName,
		&source.CompanyDomain,
		&source.IsActive,
		&source.LastPolledAt,
		&source.LastJobCount,
		&source.LastError,
		&source.CreatedAt,
		&source.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &source, nil
}
