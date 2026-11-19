package authManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type RefreshTokenRepo struct {
	postgres           *db.PostgresConn
	refreshTokensTable string
}

func NewRefreshTokenRepo(postgres *db.PostgresConn) (*RefreshTokenRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new refresh token repo: %w", config.ErrMissingDependency)
	}

	return &RefreshTokenRepo{
		postgres:           postgres,
		refreshTokensTable: config.GetTableName(config.TableRefreshTokens),
	}, nil
}

func (r *RefreshTokenRepo) Upsert(
	ctx context.Context,
	token *authModel.RefreshToken,
) (*authModel.RefreshToken, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.refreshTokensTable+` (
			id, user_id, family_id, token_hash, expires_at, revoked_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			expires_at = EXCLUDED.expires_at,
			revoked_at = EXCLUDED.revoked_at
	`, token.ID, token.UserID, token.FamilyID, token.TokenHash,
		token.ExpiresAt, token.RevokedAt, token.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert refresh token: %w", err)
	}

	return token, nil
}

func (r *RefreshTokenRepo) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*authModel.RefreshToken, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT id, user_id, family_id, token_hash, expires_at, revoked_at, created_at
		FROM `+r.refreshTokensTable+`
		WHERE token_hash = $1
		LIMIT 1
	`, tokenHash)

	return utils.ScanOptionalRow(row, scanRefreshToken, "find refresh token")
}

func (r *RefreshTokenRepo) MarkRevoked(
	ctx context.Context,
	tokenID string,
	revokedAt time.Time,
) (bool, error) {
	revoked, err := r.markRevoked(ctx, `id = $1`, tokenID, revokedAt)
	if err != nil {
		return false, fmt.Errorf("revoke refresh token: %w", err)
	}

	return revoked == 1, nil
}

func (r *RefreshTokenRepo) MarkFamilyRevoked(
	ctx context.Context,
	familyID string,
	revokedAt time.Time,
) error {
	if _, err := r.markRevoked(ctx, `family_id = $1`, familyID, revokedAt); err != nil {
		return fmt.Errorf("revoke refresh token family: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepo) MarkUserRevoked(
	ctx context.Context,
	userID string,
	revokedAt time.Time,
) error {
	if _, err := r.markRevoked(ctx, `user_id = $1`, userID, revokedAt); err != nil {
		return fmt.Errorf("revoke refresh tokens of user: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepo) markRevoked(
	ctx context.Context,
	condition string,
	argument string,
	revokedAt time.Time,
) (int64, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.refreshTokensTable+`
		SET revoked_at = $2
		WHERE `+condition+` AND revoked_at IS NULL
	`, argument, revokedAt)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

func scanRefreshToken(row utils.RowScanner) (*authModel.RefreshToken, error) {
	token := authModel.RefreshToken{}

	err := row.Scan(
		&token.ID, &token.UserID, &token.FamilyID, &token.TokenHash,
		&token.ExpiresAt, &token.RevokedAt, &token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RefreshTokenRepo) DeleteExpiredBefore(ctx context.Context, cutoff time.Time) (int, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		DELETE FROM `+r.refreshTokensTable+` WHERE expires_at < $1
	`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}

	return int(tag.RowsAffected()), nil
}
