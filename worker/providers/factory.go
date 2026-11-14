package workerProviders

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	asynqProvider "github.com/AdityaSinghRajawat/autoreach-backend/worker/providers/asynq"
)

func NewWorkerStrategy() (workerConfigs.IWorkerStrategy, error) {
	provider := workerConfigs.ProviderName(config.GetWorkerMessageProvider())
	switch provider {
	case workerConfigs.ProviderAsynq:
		return asynqProvider.NewStrategy(), nil
	default:
		return nil, fmt.Errorf(
			"new worker strategy: %w: %q",
			config.ErrUnknownWorkerProvider,
			provider,
		)
	}
}
