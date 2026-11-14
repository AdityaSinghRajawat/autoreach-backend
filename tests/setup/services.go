package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/mocks"
)

const (
	testDatabaseSuffix  = "_test"
	requireServicesFlag = "TEST_REQUIRE_SERVICES"
	setupTimeout        = 60 * time.Second
)

var (
	dependenciesOnce sync.Once
	dependenciesI    *dependencies.Dependencies
	dependenciesErr  error
	embedderI        *mocks.Embedder
)

func Dependencies(t *testing.T) *dependencies.Dependencies {
	t.Helper()
	logger := Config(t)
	require.True(
		t,
		strings.HasSuffix(config.GetDBName(), testDatabaseSuffix),
		"refusing to run against database %q: integration tests truncate tables and need a name ending in %q",
		config.GetDBName(),
		testDatabaseSuffix,
	)

	dependenciesOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
		defer cancel()

		dependenciesI, dependenciesErr = dependencies.NewDependencies(
			ctx,
			logger,
			config.GetDBMaxConns(),
		)
		if dependenciesErr != nil {
			return
		}

		embedderI = mocks.NewEmbedder(config.GetOpenAIEmbeddingDimensions())

		dependenciesI.Embedder = embedderI
		if dependenciesErr = dependenciesI.Wire(); dependenciesErr != nil {
			return
		}

		dependenciesErr = rebuildSchema(ctx, dependenciesI, RepositoryRoot(t))
	})

	if dependenciesErr != nil {
		if os.Getenv(requireServicesFlag) == "1" {
			require.NoError(t, dependenciesErr, "integration services are required but unavailable")
		}

		t.Skipf("integration services unavailable: %v", dependenciesErr)
	}

	Truncate(t, dependenciesI)
	FlushRedis(t, dependenciesI)
	embedderI.Reset()

	return dependenciesI
}

func Embedder(t *testing.T) *mocks.Embedder {
	t.Helper()
	require.NotNil(t, embedderI, "call setup.Dependencies before setup.Embedder")

	return embedderI
}

func rebuildSchema(ctx context.Context, deps *dependencies.Dependencies, root string) error {
	schema := config.GetDBSchema()
	executor := deps.Postgres.GetPool()

	statements := []string{
		`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`,
		`CREATE SCHEMA ` + schema,
	}
	for _, statement := range statements {
		if _, err := executor.Exec(ctx, statement); err != nil {
			return fmt.Errorf("rebuild schema: %w", err)
		}
	}

	deps.Postgres.GetPool().Reset()

	migrations, err := filepath.Glob(filepath.Join(root, "migrations", "*.up.sql"))
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}

	sort.Strings(migrations)

	for _, migration := range migrations {
		contents, err := os.ReadFile(migration)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filepath.Base(migration), err)
		}

		if _, err := deps.Postgres.GetPool().Exec(ctx, string(contents)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filepath.Base(migration), err)
		}
	}

	return nil
}

func Truncate(t *testing.T, deps *dependencies.Dependencies) {
	t.Helper()

	tables := config.GetDeclaredTables()
	if len(tables) == 0 {
		return
	}

	names := make([]string, 0, len(tables))
	for _, table := range tables {
		names = append(names, config.GetTableName(table))
	}

	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
	defer cancel()

	_, err := deps.Postgres.GetPool().
		Exec(ctx, `TRUNCATE TABLE `+strings.Join(names, ", ")+` RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func FlushRedis(t *testing.T, deps *dependencies.Dependencies) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
	defer cancel()

	require.NoError(t, deps.Redis.GetClient().FlushDB(ctx).Err())
}
