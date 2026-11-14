package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

const enrichmentRequestColumns = `id, job_id, contact_id, provider, status, external_request_id,
	credits_used, failure_reason, completed_at, created_at, updated_at`

type EnrichmentRequestRepo struct {
	postgres                *db.PostgresConn
	enrichmentRequestsTable string
}

func NewEnrichmentRequestRepo(postgres *db.PostgresConn) (*EnrichmentRequestRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new enrichment request repo: %w", config.ErrMissingDependency)
	}

	return &EnrichmentRequestRepo{
		postgres:                postgres,
		enrichmentRequestsTable: config.GetTableName(config.TableEnrichmentRequests),
	}, nil
}

func (r *EnrichmentRequestRepo) Upsert(
	ctx context.Context,
	request *enrichmentModel.EnrichmentRequest,
) (*enrichmentModel.EnrichmentRequest, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.enrichmentRequestsTable+` (`+enrichmentRequestColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			credits_used = EXCLUDED.credits_used,
			failure_reason = EXCLUDED.failure_reason,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
	`,
		request.ID, request.JobID, request.ContactID, request.Provider, string(request.Status),
		request.ExternalRequestID, request.CreditsUsed, request.FailureReason,
		request.CompletedAt, request.CreatedAt, request.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert enrichment request: %w", err)
	}

	return request, nil
}

func (r *EnrichmentRequestRepo) FindPendingByExternalID(
	ctx context.Context,
	provider, externalRequestID string,
) (*enrichmentModel.EnrichmentRequest, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+enrichmentRequestColumns+`
		FROM `+r.enrichmentRequestsTable+`
		WHERE provider = $1 AND external_request_id = $2 AND status = 'PENDING'
		LIMIT 1
	`, provider, externalRequestID)

	return db.FindOne(row, scanEnrichmentRequest, "find pending enrichment request")
}

func scanEnrichmentRequest(row db.RowScanner) (*enrichmentModel.EnrichmentRequest, error) {
	request := enrichmentModel.EnrichmentRequest{}

	err := row.Scan(
		&request.ID, &request.JobID, &request.ContactID, &request.Provider, &request.Status,
		&request.ExternalRequestID, &request.CreditsUsed, &request.FailureReason,
		&request.CompletedAt, &request.CreatedAt, &request.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &request, nil
}
