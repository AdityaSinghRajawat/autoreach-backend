package dependencies

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	enrichmentProviders "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type Dependencies struct {
	Logger              *slog.Logger
	Postgres            *db.PostgresConn
	Redis               *db.RedisConn
	TransactionManager  db.TransactionManager
	RateLimiter         utils.Limiter
	RedisUtil           *utils.RedisUtil
	WorkerStrategy      workerConfigs.IWorkerStrategy
	WorkerClient        workerConfigs.IWorkerClient
	Embedder            sharedTypes.Embedder
	FileStore           sharedTypes.FileStore
	EnrichmentProviders []enrichmentProviders.Provider
	EmailVerifier       *utils.EmailVerifier
	Repositories        Repositories
	Services            Services
	Handlers            Handlers
}

func NewDependencies(
	ctx context.Context,
	logger *slog.Logger,
	postgresMaxConns int,
) (*Dependencies, error) {
	if logger == nil {
		return nil, fmt.Errorf("new dependencies: %w", config.ErrMissingDependency)
	}

	dependencies := &Dependencies{Logger: logger}
	if err := dependencies.initClients(ctx, postgresMaxConns); err != nil {
		dependencies.Close()

		return nil, err
	}

	if err := dependencies.initIntegrations(); err != nil {
		dependencies.Close()

		return nil, err
	}

	if err := dependencies.Wire(); err != nil {
		dependencies.Close()

		return nil, err
	}

	return dependencies, nil
}

func (d *Dependencies) Wire() error {
	if d.Logger == nil || d.Postgres == nil || d.TransactionManager == nil ||
		d.RateLimiter == nil ||
		d.WorkerClient == nil ||
		d.Embedder == nil {
		return fmt.Errorf("wire dependencies: %w", config.ErrMissingDependency)
	}

	if err := d.initRepositories(); err != nil {
		return err
	}

	if err := d.initServices(); err != nil {
		return err
	}

	return d.initHandlers()
}

func (d *Dependencies) Close() {
	if d.WorkerClient != nil {
		if err := d.WorkerClient.Close(); err != nil {
			utils.LogErrorWithoutCtx(err)
		}
	}

	if d.Postgres != nil {
		d.Postgres.Close()
	}

	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			utils.LogErrorWithoutCtx(err)
		}
	}
}
