package outreachManagement

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

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
	_, err := r.InsertIfAbsent(ctx, entry)

	return err
}

// InsertIfAbsent reports whether the entry was added; false means the same entry already existed.
func (r *SuppressionRepo) InsertIfAbsent(
	ctx context.Context,
	entry *outreachModel.SuppressionEntry,
) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.suppressionsTable+` (id, candidate_id, entry_type, value, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`, entry.ID, entry.CandidateID, string(entry.EntryType), strings.ToLower(entry.Value),
		string(entry.Reason), entry.CreatedAt)
	if err != nil {
		return false, fmt.Errorf("insert suppression entry: %w", err)
	}

	return tag.RowsAffected() == 1, nil
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

const suppressionColumns = `id, candidate_id, entry_type, value, reason, created_at`

func (r *SuppressionRepo) List(
	ctx context.Context,
	query, candidateID string,
	limit, offset int,
) ([]*outreachModel.SuppressionEntry, int, error) {
	filter := `WHERE ($1 = '' OR value LIKE '%' || LOWER($1) || '%')
		AND ($2 = '' OR candidate_id::text = $2)`

	var total int
	if err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM `+r.suppressionsTable+` `+filter, query, candidateID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count suppression entries: %w", err)
	}

	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+suppressionColumns+` FROM `+r.suppressionsTable+` `+filter+`
		ORDER BY created_at DESC, id LIMIT $3 OFFSET $4
	`, query, candidateID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list suppression entries: %w", err)
	}

	entries, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*outreachModel.SuppressionEntry, error) {
			entry := outreachModel.SuppressionEntry{}
			err := row.Scan(
				&entry.ID,
				&entry.CandidateID,
				&entry.EntryType,
				&entry.Value,
				&entry.Reason,
				&entry.CreatedAt,
			)

			return &entry, err
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list suppression entries: %w", err)
	}

	return entries, total, nil
}

func (r *SuppressionRepo) Delete(ctx context.Context, entryID string) (bool, error) {
	tag, err := r.postgres.Executor(ctx).
		Exec(ctx, `DELETE FROM `+r.suppressionsTable+` WHERE id = $1`, entryID)
	if err != nil {
		return false, fmt.Errorf("delete suppression entry: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
