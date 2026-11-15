package outreachManagement

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type SuppressionRepo struct {
	postgres          *db.PostgresConn
	suppressionsTable string
}

func NewSuppressionRepo(postgres *db.PostgresConn) (*SuppressionRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new suppression repo: %w", config.ErrMissingDependency)
	}

	return &SuppressionRepo{
		postgres:          postgres,
		suppressionsTable: config.GetTableName(config.TableSuppressionEntries),
	}, nil
}

func (r *SuppressionRepo) Insert(ctx context.Context, entry *outreachModel.SuppressionEntry) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.suppressionsTable+` (id, candidate_id, entry_type, value, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`, entry.ID, entry.CandidateID, string(entry.EntryType), strings.ToLower(entry.Value),
		string(entry.Reason), entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert suppression entry: %w", err)
	}

	return nil
}

func (r *SuppressionRepo) IsSuppressed(
	ctx context.Context,
	candidateID, address string,
) (bool, error) {
	address = strings.ToLower(address)
	_, domain, _ := strings.Cut(address, "@")

	var suppressed bool

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM `+r.suppressionsTable+`
			WHERE (candidate_id IS NULL OR candidate_id = $1)
				AND ((entry_type = 'EMAIL' AND value = $2) OR (entry_type = 'DOMAIN' AND value = $3)))
	`, candidateID, address, domain).Scan(&suppressed)
	if err != nil {
		return false, fmt.Errorf("check suppression: %w", err)
	}

	return suppressed, nil
}
