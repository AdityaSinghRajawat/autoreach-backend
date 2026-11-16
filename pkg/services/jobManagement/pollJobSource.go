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

// PollJobSource ingests a board's open jobs and closes the ones it no longer lists; for boards that
// may not list every job, only jobs unseen for JOB_SOURCE_STALE_AFTER are closed. A failed fetch is
// recorded on the source and closes nothing.
func (s *JobService) PollJobSource(
	ctx context.Context,
	sourceID string,
) (*jobTypes.PollJobSourceResult, *config.CustomError) {
	source, customErr := s.findJobSource(ctx, sourceID)
	if customErr != nil {
		return nil, customErr
	}

	board := jobBoards.Find(s.jobBoards, string(source.BoardType))
	if board == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", config.ErrUnsupportedJobBoard),
			config.GetErrCodeBadRequest(),
		)
	}

	decision, err := s.rateLimiter.Allow(
		ctx,
		config.GetATSRateLimitPolicy(),
		string(source.BoardType),
		1,
	)
	if err == nil && !decision.Allowed {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", config.ErrATSRateLimited),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	knownExternalIDs, err := s.jobSourceRepo.ListActiveExternalIDs(ctx, source.ID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	pollStartedAt := utils.GetCurrTimeStamp()

	listing, err := board.ListJobs(ctx, source, knownExternalIDs)
	if err != nil {
		source.RecordPoll(0, err)

		if customErr := s.recordPoll(ctx, source); customErr != nil {
			return nil, customErr
		}

		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			utils.ClassifyUpstreamError(err),
		)
	}

	s.resolveCompanyDomains(ctx, listing.Jobs)

	result, customErr := s.ingestBoardJobs(ctx, source, listing.Jobs)
	if customErr != nil {
		return nil, customErr
	}

	closeBefore := pollStartedAt
	if !listing.Complete {
		closeBefore = pollStartedAt.Add(-config.GetJobSourceStaleAfter())
	}

	if result.Closed, err = s.jobSourceRepo.CloseUnseen(
		ctx,
		source.ID,
		closeBefore,
		utils.GetCurrTimeStamp(),
	); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	source.RecordPoll(result.Fetched, nil)

	return result, s.recordPoll(ctx, source)
}

// resolveCompanyDomains fills in the domain of jobs from boards that list many companies; jobs
// whose domain stays unknown fail validation and are skipped.
func (s *JobService) resolveCompanyDomains(ctx context.Context, boardJobs []jobTypes.BoardJob) {
	companyNames := []string{}

	for _, boardJob := range boardJobs {
		if boardJob.CompanyName != "" && boardJob.CompanyDomain == "" && !boardJob.DetailsSkipped {
			companyNames = append(companyNames, boardJob.CompanyName)
		}
	}

	if len(companyNames) == 0 {
		return
	}

	domains, err := s.domainResolver.ResolveCompanyDomains(ctx, companyNames)
	if err != nil {
		utils.LogWarn(ctx, "company domain resolution failed", "error", err.Error())

		return
	}

	for index := range boardJobs {
		if domain, found := domains[boardJobs[index].CompanyName]; found {
			boardJobs[index].CompanyDomain = domain
		}
	}
}

func (s *JobService) ingestBoardJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	boardJobs []jobTypes.BoardJob,
) (*jobTypes.PollJobSourceResult, *config.CustomError) {
	result := &jobTypes.PollJobSourceResult{SourceID: source.ID, Fetched: len(boardJobs)}
	valid := make([]*jobModel.JobPosting, 0, len(boardJobs))
	unchangedExternalIDs := []string{}

	for _, boardJob := range boardJobs {
		if boardJob.DetailsSkipped {
			unchangedExternalIDs = append(unchangedExternalIDs, boardJob.ExternalID)

			continue
		}

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

	unchanged, err := s.jobSourceRepo.MarkSeen(
		ctx,
		source.ID,
		unchangedExternalIDs,
		utils.GetCurrTimeStamp(),
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	result.Unchanged = unchanged

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

func (s *JobService) recordPoll(
	ctx context.Context,
	source *jobModel.JobSource,
) *config.CustomError {
	source.UpdateTimestamp()

	if err := s.jobSourceRepo.RecordPoll(ctx, source); err != nil {
		return config.NewCustomError(
			fmt.Errorf("poll job source: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}
