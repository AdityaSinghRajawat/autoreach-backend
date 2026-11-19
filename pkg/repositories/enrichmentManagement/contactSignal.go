package enrichmentManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type ContactSignalRepo struct {
	postgres     *db.PostgresConn
	signalsTable string
}

func NewContactSignalRepo(postgres *db.PostgresConn) (*ContactSignalRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new contact signal repo: %w", config.ErrMissingDependency)
	}

	return &ContactSignalRepo{
		postgres:     postgres,
		signalsTable: config.GetTableName(config.TableContactSignals),
	}, nil
}

const contactSignalColumns = `id, contact_id, candidate_id, signal_type, address, occurred_at,
	applied_at, created_at, updated_at`

// Upsert keeps one signal per candidate, so repeating an observation never counts twice.
func (r *ContactSignalRepo) Upsert(
	ctx context.Context,
	signal *enrichmentModel.ContactSignal,
) (*enrichmentModel.ContactSignal, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.signalsTable+` (`+contactSignalColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT ON CONSTRAINT unique_contact_signal_reporter
		DO UPDATE SET occurred_at = EXCLUDED.occurred_at, updated_at = EXCLUDED.updated_at
	`, signal.ID, signal.ContactID, signal.CandidateID, string(signal.SignalType), signal.Address,
		signal.OccurredAt, signal.AppliedAt, signal.CreatedAt, signal.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert contact signal: %w", err)
	}

	return signal, nil
}

func (r *ContactSignalRepo) CountReporters(
	ctx context.Context,
	signal *enrichmentModel.ContactSignal,
) (int, error) {
	var reporters int

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM `+r.signalsTable+`
		WHERE contact_id = $1 AND signal_type = $2 AND address = $3
	`, signal.ContactID, string(signal.SignalType), signal.Address).Scan(&reporters)
	if err != nil {
		return 0, fmt.Errorf("count contact signal reporters: %w", err)
	}

	return reporters, nil
}

// MarkApplied stamps every report of the same observation, since they were applied together.
func (r *ContactSignalRepo) MarkApplied(
	ctx context.Context,
	signal *enrichmentModel.ContactSignal,
	now time.Time,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.signalsTable+` SET applied_at = $4, updated_at = $4
		WHERE contact_id = $1 AND signal_type = $2 AND address = $3 AND applied_at IS NULL
	`, signal.ContactID, string(signal.SignalType), signal.Address, now)
	if err != nil {
		return fmt.Errorf("mark contact signal applied: %w", err)
	}

	return nil
}

func (r *ContactSignalRepo) ListByContact(
	ctx context.Context,
	contactID string,
) ([]*enrichmentModel.ContactSignal, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+contactSignalColumns+` FROM `+r.signalsTable+`
		WHERE contact_id = $1 ORDER BY occurred_at DESC, id
	`, contactID)
	if err != nil {
		return nil, fmt.Errorf("list contact signals: %w", err)
	}

	signals, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*enrichmentModel.ContactSignal, error) {
			return scanContactSignal(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list contact signals: %w", err)
	}

	return signals, nil
}

func (r *ContactSignalRepo) FindByContact(
	ctx context.Context,
	contactID, signalID string,
) (*enrichmentModel.ContactSignal, error) {
	return utils.ScanOptionalRow(r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+contactSignalColumns+` FROM `+r.signalsTable+` WHERE id = $1 AND contact_id = $2
	`, signalID, contactID), scanContactSignal, "find contact signal")
}

func scanContactSignal(row utils.RowScanner) (*enrichmentModel.ContactSignal, error) {
	signal := enrichmentModel.ContactSignal{}
	err := row.Scan(
		&signal.ID,
		&signal.ContactID,
		&signal.CandidateID,
		&signal.SignalType,
		&signal.Address,
		&signal.OccurredAt,
		&signal.AppliedAt,
		&signal.CreatedAt,
		&signal.UpdatedAt,
	)

	return &signal, err
}
