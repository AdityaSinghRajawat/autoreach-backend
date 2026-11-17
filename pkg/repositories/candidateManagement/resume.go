package candidateManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

const resumeColumns = `id, candidate_id, file_name, storage_key, content_type, size_bytes,
	content_sha256, raw_text, parsed_metadata, parse_status::text, failure_reason,
	embedding_model, is_active, created_at, updated_at`

type ResumeRepo struct {
	postgres         *db.PostgresConn
	resumesTable     string
	candidatesTable  string
	jobPostingsTable string
}

func NewResumeRepo(postgres *db.PostgresConn) (*ResumeRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new resume repo: %w", config.ErrMissingDependency)
	}

	return &ResumeRepo{
		postgres:         postgres,
		resumesTable:     config.GetTableName(config.TableCandidateResumes),
		candidatesTable:  config.GetTableName(config.TableCandidates),
		jobPostingsTable: config.GetTableName(config.TableJobPostings),
	}, nil
}

func (r *ResumeRepo) Upsert(
	ctx context.Context,
	resume *candidateModel.Resume,
) (*candidateModel.Resume, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.resumesTable+` (
			id, candidate_id, file_name, storage_key, content_type, size_bytes, content_sha256,
			raw_text, parsed_metadata, parse_status, failure_reason, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::resume_status, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			file_name = EXCLUDED.file_name,
			raw_text = EXCLUDED.raw_text,
			parsed_metadata = EXCLUDED.parsed_metadata,
			parse_status = EXCLUDED.parse_status,
			failure_reason = EXCLUDED.failure_reason,
			updated_at = EXCLUDED.updated_at
	`,
		resume.ID, resume.CandidateID, resume.FileName, resume.StorageKey, resume.ContentType,
		resume.SizeBytes, resume.ContentSHA256, resume.RawText, string(resume.ParsedMetadata),
		string(resume.Status), resume.FailureReason, resume.CreatedAt, resume.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert resume: %w", err)
	}

	return resume, nil
}

func (r *ResumeRepo) FindByID(
	ctx context.Context,
	resumeID string,
) (*candidateModel.Resume, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+resumeColumns+` FROM `+r.resumesTable+` WHERE id = $1 LIMIT 1
	`, resumeID)

	return db.FindOne(row, scanResume, "find resume")
}

func (r *ResumeRepo) ListByCandidate(
	ctx context.Context,
	candidateID string,
) ([]*candidateModel.Resume, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+resumeColumns+` FROM `+r.resumesTable+`
		WHERE candidate_id = $1
		ORDER BY created_at DESC
	`, candidateID)
	if err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}

	resumes, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*candidateModel.Resume, error) {
			return scanResume(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}

	return resumes, nil
}

func (r *ResumeRepo) ListStalePendingIDs(
	ctx context.Context,
	staleBefore time.Time,
	limit int,
) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT id::text FROM `+r.resumesTable+`
		WHERE parse_status IN (`+sqlResumeStatusUploaded+`, `+sqlResumeStatusParsed+`) AND updated_at < $1
		ORDER BY updated_at
		LIMIT $2
	`, staleBefore, limit)
	if err != nil {
		return nil, fmt.Errorf("list stale pending resumes: %w", err)
	}

	resumeIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list stale pending resumes: %w", err)
	}

	return resumeIDs, nil
}

func (r *ResumeRepo) UpdateEmbedding(
	ctx context.Context,
	resumeID string,
	embedding []float32,
	model string,
	embeddedAt time.Time,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.resumesTable+`
		SET embedding = $2::vector, embedding_model = $3, parse_status = `+sqlResumeStatusEmbedded+`,
			failure_reason = NULL, updated_at = $4
		WHERE id = $1
	`, resumeID, db.FormatVector(embedding), model, embeddedAt)
	if err != nil {
		return fmt.Errorf("update resume embedding: %w", err)
	}

	return nil
}

func (r *ResumeRepo) Activate(
	ctx context.Context,
	candidateID, resumeID string,
	activatedAt time.Time,
) (bool, error) {
	executor := r.postgres.Executor(ctx)

	_, err := executor.Exec(ctx, `
		UPDATE `+r.resumesTable+`
		SET is_active = FALSE, updated_at = $3
		WHERE candidate_id = $1 AND is_active AND id <> $2
	`, candidateID, resumeID, activatedAt)
	if err != nil {
		return false, fmt.Errorf("deactivate other resumes: %w", err)
	}

	tag, err := executor.Exec(ctx, `
		UPDATE `+r.resumesTable+`
		SET is_active = TRUE, updated_at = $3
		WHERE id = $2 AND candidate_id = $1 AND parse_status = `+sqlResumeStatusEmbedded+`
	`, candidateID, resumeID, activatedAt)
	if err != nil {
		return false, fmt.Errorf("activate resume: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *ResumeRepo) Delete(ctx context.Context, candidateID, resumeID string) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		DELETE FROM `+r.resumesTable+` WHERE id = $1 AND candidate_id = $2
	`, resumeID, candidateID)
	if err != nil {
		return false, fmt.Errorf("delete resume: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *ResumeRepo) FindEmbedding(
	ctx context.Context,
	resumeID string,
) (*matchingTypes.ResumeEmbedding, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT id, candidate_id, embedding::text, is_active
		FROM `+r.resumesTable+`
		WHERE id = $1 AND embedding IS NOT NULL
		LIMIT 1
	`, resumeID)

	return db.FindOne(row, scanResumeEmbedding, "find resume embedding")
}

func (r *ResumeRepo) ListActiveSimilaritiesToJob(
	ctx context.Context,
	jobID string,
) ([]matchingTypes.ResumeSimilarity, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT resume.id, resume.candidate_id, candidate.min_match_score::float8,
			(1 - (resume.embedding <=> job.embedding))::float8,
			candidate.auto_outreach_enabled
		FROM `+r.resumesTable+` AS resume
		JOIN `+r.candidatesTable+` AS candidate ON candidate.id = resume.candidate_id
		JOIN `+r.jobPostingsTable+` AS job ON job.id = $1
		WHERE resume.is_active
			AND resume.embedding IS NOT NULL
			AND job.embedding IS NOT NULL
			AND job.status = `+sqlJobStatusActive+`
	`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list resume similarities to job: %w", err)
	}

	similarities, err := pgx.CollectRows(rows, pgx.RowToStructByPos[matchingTypes.ResumeSimilarity])
	if err != nil {
		return nil, fmt.Errorf("list resume similarities to job: %w", err)
	}

	return similarities, nil
}

func scanResume(row db.RowScanner) (*candidateModel.Resume, error) {
	resume := candidateModel.Resume{}

	var parsedMetadata []byte

	err := row.Scan(
		&resume.ID, &resume.CandidateID, &resume.FileName, &resume.StorageKey,
		&resume.ContentType, &resume.SizeBytes, &resume.ContentSHA256, &resume.RawText,
		&parsedMetadata, &resume.Status, &resume.FailureReason, &resume.EmbeddingModel,
		&resume.IsActive, &resume.CreatedAt, &resume.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	resume.ParsedMetadata = parsedMetadata

	return &resume, nil
}

func scanResumeEmbedding(row db.RowScanner) (*matchingTypes.ResumeEmbedding, error) {
	embedding := matchingTypes.ResumeEmbedding{}

	err := row.Scan(
		&embedding.ResumeID, &embedding.CandidateID, &embedding.Vector, &embedding.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &embedding, nil
}

func (r *ResumeRepo) ListStorageKeys(ctx context.Context, candidateID string) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT storage_key FROM `+r.resumesTable+` WHERE candidate_id = $1
	`, candidateID)
	if err != nil {
		return nil, fmt.Errorf("list resume storage keys: %w", err)
	}

	keys, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list resume storage keys: %w", err)
	}

	return keys, nil
}
