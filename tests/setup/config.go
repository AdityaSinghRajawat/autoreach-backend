package setup

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	configOnce sync.Once
	configErr  error
	loggerI    *slog.Logger
)

func Config(t *testing.T) *slog.Logger {
	t.Helper()
	configOnce.Do(func() {
		root, err := repositoryRoot()
		if err != nil {
			configErr = err

			return
		}

		if err := godotenv.Overload(filepath.Join(root, ".env.test")); err != nil {
			configErr = err

			return
		}

		config.Init()

		loggerI = utils.InitLogger("test")
	})
	require.NoError(t, configErr)

	return loggerI
}

func RepositoryRoot(t *testing.T) string {
	t.Helper()

	root, err := repositoryRoot()
	require.NoError(t, err)

	return root
}

func repositoryRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", os.ErrNotExist
		}

		directory = parent
	}
}
