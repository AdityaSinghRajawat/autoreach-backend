package workerTasks

import (
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type taskPayload interface {
	Validate() error
}

func decodePayload(
	message workerConfigs.Message,
	payload taskPayload,
) *workerConfigs.WorkerMessageResult {
	if err := json.Unmarshal(message.Payload, payload); err != nil {
		return workerConfigs.NewWorkerFailureResult(
			fmt.Errorf("decode %s payload: %w", message.TaskID, err),
			config.GetErrCodeBadRequest(),
			false,
		)
	}

	if err := payload.Validate(); err != nil {
		return workerConfigs.NewWorkerFailureResult(err, config.GetErrCodeValidation(), false)
	}

	return nil
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
