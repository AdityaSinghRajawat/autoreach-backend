package shared

import (
	"context"
	"time"

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

type RateLimitDecision struct {
	Allowed    bool
	Remaining  int
	NextUnitIn time.Duration
}

// A cost of 0 checks the allowance without spending it.
type Limiter interface {
	Allow(
		ctx context.Context,
		policy config.RateLimitPolicy,
		subject string,
		cost int,
	) (RateLimitDecision, error)
}

type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (string, error)
	ChatModel() string
}

type CompanyDomainResolver interface {
	ResolveCompanyDomains(ctx context.Context, companyNames []string) (map[string]string, error)
}
