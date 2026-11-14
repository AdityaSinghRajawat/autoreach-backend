package authManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
)

const apiClientColumns = `id, name, key_prefix, key_hash, scopes, last_used_at, revoked_at,
	created_at, updated_at`

type APIClientRepo struct {
	postgres        *db.PostgresConn
	apiClientsTable string
}

func NewAPIClientRepo(postgres *db.PostgresConn) (*APIClientRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new api client repo: %w", config.ErrMissingDependency)
	}

	return &APIClientRepo{
		postgres:        postgres,
		apiClientsTable: config.GetTableName(config.TableAPIClients),
	}, nil
}

func (r *APIClientRepo) Upsert(
	ctx context.Context,
	client *authModel.APIClient,
) (*authModel.APIClient, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.apiClientsTable+` (`+apiClientColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			scopes = EXCLUDED.scopes,
			last_used_at = EXCLUDED.last_used_at,
			revoked_at = EXCLUDED.revoked_at,
			updated_at = EXCLUDED.updated_at
	`, client.ID, client.Name, client.KeyPrefix, client.KeyHash, client.Scopes,
		client.LastUsedAt, client.RevokedAt, client.CreatedAt, client.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert api client: %w", err)
	}

	return client, nil
}

func (r *APIClientRepo) FindActiveByPrefix(
	ctx context.Context,
	keyPrefix string,
) (*authModel.APIClient, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+apiClientColumns+` FROM `+r.apiClientsTable+`
		WHERE key_prefix = $1 AND revoked_at IS NULL
		LIMIT 1
	`, keyPrefix)

	return db.FindOne(row, scanAPIClient, "find api client by prefix")
}

func (r *APIClientRepo) List(ctx context.Context) ([]*authModel.APIClient, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+apiClientColumns+` FROM `+r.apiClientsTable+` ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list api clients: %w", err)
	}

	clients, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*authModel.APIClient, error) {
			return scanAPIClient(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list api clients: %w", err)
	}

	return clients, nil
}

func (r *APIClientRepo) MarkRevoked(
	ctx context.Context,
	clientID string,
	revokedAt time.Time,
) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.apiClientsTable+`
		SET revoked_at = $2, updated_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`, clientID, revokedAt)
	if err != nil {
		return false, fmt.Errorf("revoke api client: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *APIClientRepo) UpdateLastUsedAt(
	ctx context.Context,
	clientID string,
	usedAt time.Time,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.apiClientsTable+` SET last_used_at = $2 WHERE id = $1
	`, clientID, usedAt)
	if err != nil {
		return fmt.Errorf("update api client last used: %w", err)
	}

	return nil
}

func scanAPIClient(row db.RowScanner) (*authModel.APIClient, error) {
	client := authModel.APIClient{}

	err := row.Scan(
		&client.ID, &client.Name, &client.KeyPrefix, &client.KeyHash, &client.Scopes,
		&client.LastUsedAt, &client.RevokedAt, &client.CreatedAt, &client.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &client, nil
}
