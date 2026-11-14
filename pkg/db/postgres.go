package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type PostgresConn struct {
	pool *pgxpool.Pool
}

func ConnectPostgres(ctx context.Context, maxConns int) (*PostgresConn, error) {
	poolConfig, err := pgxpool.ParseConfig(config.GetDBDSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolConfig.MaxConns = int32(maxConns)
	poolConfig.MaxConnLifetime = config.GetDBMaxConnLifetime()
	poolConfig.HealthCheckPeriod = config.GetDBHealthCheckPeriod()
	poolConfig.ConnConfig.ConnectTimeout = config.GetDBConnectTimeout()
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = milliseconds(
		config.GetDBStatementTimeout().Milliseconds(),
	)
	poolConfig.ConnConfig.RuntimeParams["lock_timeout"] = milliseconds(
		config.GetDBLockTimeout().Milliseconds(),
	)
	poolConfig.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = milliseconds(
		config.GetDBIdleInTransactionTimeout().Milliseconds(),
	)
	poolConfig.ConnConfig.RuntimeParams["application_name"] = config.GetServiceName()
	poolConfig.AfterConnect = decodeTimestampsAsUTC

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	conn := &PostgresConn{pool: pool}
	if err := conn.Ping(ctx); err != nil {
		pool.Close()

		return nil, err
	}

	return conn, nil
}

func milliseconds(value int64) string {
	return strconv.FormatInt(value, 10)
}

func (c *PostgresConn) Executor(ctx context.Context) Executor {
	if transaction, open := TransactionFromContext(ctx); open {
		return transaction
	}

	return c.pool
}

func (c *PostgresConn) GetPool() *pgxpool.Pool {
	return c.pool
}

func (c *PostgresConn) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, config.GetDBConnectTimeout())
	defer cancel()

	if err := c.pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}

func (c *PostgresConn) Close() {
	c.pool.Close()
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func IsUniqueViolation(err error) bool {
	var postgresErr *pgconn.PgError

	return errors.As(err, &postgresErr) && postgresErr.Code == config.GetUniqueViolationCode()
}

func decodeTimestampsAsUTC(_ context.Context, conn *pgx.Conn) error {
	conn.TypeMap().RegisterType(&pgtype.Type{
		Name:  "timestamptz",
		OID:   pgtype.TimestamptzOID,
		Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
	})

	return nil
}
