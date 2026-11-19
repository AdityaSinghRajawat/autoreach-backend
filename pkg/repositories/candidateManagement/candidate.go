package candidateManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

const candidateColumns = `id, user_id, email, full_name, linkedin_profile_url,
	min_match_score::float8, auto_apply_enabled, auto_outreach_enabled, timezone, settings,
	created_at, updated_at`

type CandidateRepo struct {
	postgres        *db.PostgresConn
	candidatesTable string
}

func NewCandidateRepo(postgres *db.PostgresConn) (*CandidateRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new candidate repo: %w", config.ErrMissingDependency)
	}

	return &CandidateRepo{
		postgres:        postgres,
		candidatesTable: config.GetTableName(config.TableCandidates),
	}, nil
}

func (r *CandidateRepo) Upsert(
	ctx context.Context,
	candidate *candidateModel.Candidate,
) (*candidateModel.Candidate, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.candidatesTable+` (
			id, user_id, email, full_name, linkedin_profile_url, min_match_score,
			auto_apply_enabled, auto_outreach_enabled, timezone, settings, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			full_name = EXCLUDED.full_name,
			linkedin_profile_url = EXCLUDED.linkedin_profile_url,
			min_match_score = EXCLUDED.min_match_score,
			auto_apply_enabled = EXCLUDED.auto_apply_enabled,
			auto_outreach_enabled = EXCLUDED.auto_outreach_enabled,
			timezone = EXCLUDED.timezone,
			settings = EXCLUDED.settings,
			updated_at = EXCLUDED.updated_at
	`,
		candidate.ID,
		candidate.UserID,
		candidate.Email,
		candidate.FullName,
		candidate.LinkedinProfileURL,
		candidate.MinMatchScore,
		candidate.AutoApplyEnabled,
		candidate.AutoOutreachEnabled,
		candidate.Timezone,
		string(candidate.Settings),
		candidate.CreatedAt,
		candidate.UpdatedAt,
	)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, fmt.Errorf("upsert candidate: %w", config.ErrDuplicateRecord)
		}

		return nil, fmt.Errorf("upsert candidate: %w", err)
	}

	return candidate, nil
}

func (r *CandidateRepo) FindByID(
	ctx context.Context,
	candidateID string,
) (*candidateModel.Candidate, error) {
	return r.findOne(ctx, `id = $1`, candidateID, "find candidate by id")
}

func (r *CandidateRepo) FindByUserID(
	ctx context.Context,
	userID string,
) (*candidateModel.Candidate, error) {
	return r.findOne(ctx, `user_id = $1`, userID, "find candidate by user id")
}

func (r *CandidateRepo) findOne(
	ctx context.Context,
	condition string,
	argument string,
	operation string,
) (*candidateModel.Candidate, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+candidateColumns+` FROM `+r.candidatesTable+` WHERE `+condition+` LIMIT 1
	`, argument)

	return utils.ScanOptionalRow(row, scanCandidate, operation)
}

func scanCandidate(row utils.RowScanner) (*candidateModel.Candidate, error) {
	candidate := candidateModel.Candidate{}

	var settings []byte

	err := row.Scan(
		&candidate.ID,
		&candidate.UserID,
		&candidate.Email,
		&candidate.FullName,
		&candidate.LinkedinProfileURL,
		&candidate.MinMatchScore,
		&candidate.AutoApplyEnabled,
		&candidate.AutoOutreachEnabled,
		&candidate.Timezone,
		&settings,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	candidate.Settings = settings

	return &candidate, nil
}

func (r *CandidateRepo) UpdateEmailForUser(
	ctx context.Context,
	userID, email string,
	now time.Time,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.candidatesTable+` SET email = $2, updated_at = $3 WHERE user_id = $1
	`, userID, email, now)
	if db.IsUniqueViolation(err) {
		return fmt.Errorf("update candidate email: %w", config.ErrDuplicateRecord)
	}

	if err != nil {
		return fmt.Errorf("update candidate email: %w", err)
	}

	return nil
}
