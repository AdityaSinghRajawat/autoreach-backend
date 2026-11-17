package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *JobService) EmbedJobs(
	ctx context.Context,
	jobIDs []string,
) (*jobTypes.EmbedJobsResult, *config.CustomError) {
	result := &jobTypes.EmbedJobsResult{Requested: len(jobIDs)}

	sources, err := s.jobPostingRepo.ListEmbeddingSources(ctx, jobIDs)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("embed jobs: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if len(sources) == 0 {
		return result, nil
	}

	maxChars := config.GetAIEmbeddingMaxInputChars()

	inputs := make([]string, 0, len(sources))
	for _, source := range sources {
		inputs = append(inputs, source.EmbeddingText(maxChars))
	}

	embeddings, err := s.embedder.CreateEmbeddings(ctx, inputs)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("embed jobs: %w", err),
			utils.ClassifyUpstreamError(err),
		)
	}

	if len(embeddings) != len(sources) {
		return nil, config.NewCustomError(
			fmt.Errorf("embed jobs: %w", config.ErrEmbeddingCountMismatch),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	updates := make([]jobTypes.JobEmbeddingUpdate, 0, len(sources))
	for index, source := range sources {
		updates = append(updates, jobTypes.NewJobEmbeddingUpdate(source, embeddings[index]))
	}

	embedded, err := s.jobPostingRepo.UpdateEmbeddings(
		ctx,
		updates,
		s.embedder.Model(),
		utils.GetCurrTimeStamp(),
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("embed jobs: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	result.Embedded = embedded

	s.publishMatchJobTasks(ctx, sources)

	return result, nil
}

func (s *JobService) publishMatchJobTasks(
	ctx context.Context,
	sources []jobTypes.JobEmbeddingSource,
) {
	tasks := make([]utils.ConcurrentTask, 0, len(sources))
	for _, source := range sources {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish match job task",
			Run: func(ctx context.Context) error {
				_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
					TaskID:  workerConfigs.TaskMatchJob,
					Payload: workerCustomTypes.MatchJobPayload{JobID: source.ID},
				})
				if err != nil {
					utils.LogWarn(
						ctx,
						"match job task publish failed",
						"job_id", source.ID,
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
}
