package shared

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type Embedder interface {
	CreateEmbeddings(ctx context.Context, inputs []string) ([][]float32, error)
	Model() string
}

type FileStore interface {
	Save(ctx context.Context, key string, content []byte) error
	Read(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type APIKeyVerifier interface {
	VerifyAPIKey(ctx context.Context, apiKey string) (*Principal, *config.CustomError)
}
