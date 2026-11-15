package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement/jobBoards"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// PollJobSource ingests a board's open jobs and closes the ones it no longer lists. A failed fetch
// is recorded on the source and closes nothing.
func (s *JobService) PollJobSource(
	ctx context.Context,
	sourceID string,
) (*jobTypes.PollJobSourceResult, *config.CustomError) {
	source, customErr := s.findJobSource(ctx, sourceID)
	if customErr != nil {
		return nil, customErr
	}

	board := jobBoards.Find(s.jobBoards, source.ATSType)
	if board == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", config.ErrUnsupportedJobBoard),
			config.GetErrCodeBadRequest(),
		)
	}

	if decision, err := s.rateLimiter.Allow(
		ctx,
		config.GetATSRateLimitPolicy(),
		source.ATSType,
		1,
	); err == nil &&
		!decision.Allowed {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", config.ErrATSRateLimited),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	pollStartedAt := utils.GetCurrTimeStamp()

	boardJobs, err := board.ListJobs(ctx, source)
	if err != nil {
		source.RecordPoll(0, err)

		if customErr := s.saveJobSource(ctx, source, "poll job source"); customErr != nil {
			return nil, customErr
		}

		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			utils.ClassifyUpstreamError(err),
		)
	}

	result, customErr := s.ingestBoardJobs(ctx, source, boardJobs)
	if customErr != nil {
		return nil, customErr
	}

	if result.Closed, err = s.jobSourceRepo.CloseUnseen(
		ctx,
		source.ID,
		pollStartedAt,
		utils.GetCurrTimeStamp(),
	); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	source.RecordPoll(result.Fetched, nil)

	return result, s.saveJobSource(ctx, source, "poll job source")
}

func (s *JobService) ingestBoardJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	boardJobs []jobTypes.BoardJob,
) (*jobTypes.PollJobSourceResult, *config.CustomError) {
	result := &jobTypes.PollJobSourceResult{SourceID: source.ID, Fetched: len(boardJobs)}
	valid := make([]*jobModel.JobPosting, 0, len(boardJobs))

	for _, boardJob := range boardJobs {
		item := boardJob.ToIngestJobItem(source)
		job := item.ToJobPostingModel()
		job.JobSourceID = &source.ID

		if job.Validate() != nil {
			result.Skipped++

			continue
		}

		valid = append(valid, job)
	}

	for start := 0; start < len(valid); start += config.GetIngestBatchSize() {
		batch := valid[start:min(start+config.GetIngestBatchSize(), len(valid))]

		ingested, customErr := s.IngestJobs(ctx, batch)
		if customErr != nil {
			return nil, customErr
		}

		result.Ingested += ingested.IngestedCount
	}

	return result, nil
}

func (s *JobService) ScheduleJobSourcePolls(ctx context.Context) (int, *config.CustomError) {
	sourceIDs, err := s.jobSourceRepo.ListActiveIDs(ctx)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("schedule job source polls: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	for _, sourceID := range sourceIDs {
		s.publishPollJobSource(ctx, sourceID)
	}

	return len(sourceIDs), nil
}
