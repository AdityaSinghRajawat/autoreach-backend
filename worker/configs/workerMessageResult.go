package workerConfigs

import "github.com/AdityaSinghRajawat/autoreach-backend/config"

type WorkerMessageResult struct {
	Success     bool
	Err         error
	ErrCode     config.CustomErrorCode
	IsRetryable bool
}

func NewWorkerSuccessResult() *WorkerMessageResult {
	return &WorkerMessageResult{Success: true}
}

func NewWorkerFailureResult(
	err error,
	errCode config.CustomErrorCode,
	isRetryable bool,
) *WorkerMessageResult {
	return &WorkerMessageResult{
		Err:         err,
		ErrCode:     errCode,
		IsRetryable: isRetryable,
	}
}
