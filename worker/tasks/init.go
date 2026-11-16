package workerTasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

// handle decodes and validates the payload once, so each task only maps it onto its service call.
func handle[P any](
	run func(ctx context.Context, payload *P) *config.CustomError,
) workerConfigs.WorkerMessageHandler {
	return func(ctx context.Context, message workerConfigs.Message) *workerConfigs.WorkerMessageResult {
		payload := new(P)
		if err := json.Unmarshal(message.Payload, payload); err != nil {
			return workerConfigs.NewWorkerFailureResult(
				fmt.Errorf("decode %s payload: %w", message.TaskID, err),
				config.GetErrCodeBadRequest(),
				false,
			)
		}

		if err := validations.Validate(payload); err != nil {
			return workerConfigs.NewWorkerFailureResult(
				fmt.Errorf("validate %s payload: %w", message.TaskID, err),
				config.GetErrCodeValidation(),
				false,
			)
		}

		return resultFor(run(ctx, payload))
	}
}

func resultFor(customErr *config.CustomError) *workerConfigs.WorkerMessageResult {
	if customErr != nil {
		return workerConfigs.NewWorkerFailureResult(
			customErr,
			customErr.ErrCode,
			config.IsRetryableErrCode(customErr.ErrCode),
		)
	}

	return workerConfigs.NewWorkerSuccessResult()
}
