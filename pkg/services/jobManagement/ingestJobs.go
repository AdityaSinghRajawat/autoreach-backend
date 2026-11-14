package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *JobService) IngestJobs(
	ctx context.Context,
	jobs []*jobModel.JobPosting,
) (*jobTypes.IngestJobsResponse, *config.CustomError) {
	response := jobTypes.NewIngestJobsResponse()

	for index, job := range jobs {
		job.UpdateTimestamp()

		if err := job.Validate(); err != nil {
			return nil, config.NewCustomError(
				fmt.Errorf("ingest jobs: job %d: %w", index, err),
				config.GetErrCodeBadRequest(),
			)
		}
	}

	var upserted []jobTypes.UpsertedJob

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		upserted, err = s.jobPostingRepo.UpsertMany(ctx, jobs)

		return err
	})
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("ingest jobs: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	response.IngestedCount = len(upserted)
	response.QueuedTasks = s.publishEmbedTasks(ctx, upserted)

	return response, nil
}

func (s *JobService) publishEmbedTasks(
	ctx context.Context,
	upserted []jobTypes.UpsertedJob,
) []jobTypes.QueuedTask {
	queued := make([]jobTypes.QueuedTask, 0, len(upserted))
	seen := make(map[string]struct{}, len(upserted))

	for _, job := range upserted {
		if _, duplicate := seen[job.ID]; duplicate || !job.NeedsEmbedding {
			continue
		}

		seen[job.ID] = struct{}{}
		queued = append(queued, jobTypes.NewQueuedTask(
			job.ID,
			string(workerConfigs.TaskEmbedJob),
			workerConfigs.PublishStatusEnqueued,
		))
	}

	tasks := make([]utils.ConcurrentTask, 0, len(queued))
	for index := range queued {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish embed task",
			Run: func(ctx context.Context) error {
				if err := s.publishEmbedTask(ctx, queued[index].JobID); err != nil {
					queued[index].Status = workerConfigs.PublishStatusDeferred

					utils.LogWarn(
						ctx,
						"embed task publish failed, reconciler will retry",
						"job_id", queued[index].JobID,
						"error", err.Error(),
					)
				}

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(context.WithoutCancel(ctx), tasks...); err != nil {
		utils.LogError(ctx, err)
	}

	return queued
}

func (s *JobService) publishEmbedTask(ctx context.Context, jobID string) error {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskEmbedJob,
		Payload: workerCustomTypes.EmbedJobPayload{JobID: jobID},
	})

	return err
}
