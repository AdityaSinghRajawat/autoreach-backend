package workerAggregators

import workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"

func NewWorkerAggregatorRegistry() workerConfigs.WorkerAggregatorRegistry {
	return workerConfigs.WorkerAggregatorRegistry{
		workerConfigs.GroupEmbedJob: AggregateEmbedJobs,
	}
}
