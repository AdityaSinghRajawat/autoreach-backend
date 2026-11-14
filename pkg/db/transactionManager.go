package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type Executor interface {
	Exec(ctx context.Context, statement string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, statement string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, statement string, arguments ...any) pgx.Row
	SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, work func(ctx context.Context) error) error
}

type transactionContextKey struct{}

type PostgresTransactionManager struct {
	postgres *PostgresConn
}

func NewPostgresTransactionManager(postgres *PostgresConn) (*PostgresTransactionManager, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new transaction manager: %w", config.ErrMissingDependency)
	}

	return &PostgresTransactionManager{postgres: postgres}, nil
}

func TransactionFromContext(ctx context.Context) (pgx.Tx, bool) {
	transaction, open := ctx.Value(transactionContextKey{}).(pgx.Tx)

	return transaction, open
}

func (m *PostgresTransactionManager) WithTransaction(
	ctx context.Context,
	work func(ctx context.Context) error,
) (err error) {
	if _, open := TransactionFromContext(ctx); open {
		return work(ctx)
	}

	transaction, err := m.postgres.GetPool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}

		rollbackErr := transaction.Rollback(context.WithoutCancel(ctx))
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr))
		}
	}()

	if err = work(context.WithValue(ctx, transactionContextKey{}, transaction)); err != nil {
		return err
	}

	if err = transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	committed = true

	return nil
}
