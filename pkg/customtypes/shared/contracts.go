package shared

import "context"

type Embedder interface {
	CreateEmbeddings(ctx context.Context, inputs []string) ([][]float32, error)
	Model() string
}
