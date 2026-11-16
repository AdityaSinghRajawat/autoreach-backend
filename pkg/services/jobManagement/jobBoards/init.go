package jobBoards

import (
	"context"
	"net/http"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	ashbyIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/ashby"
	greenhouseIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/greenhouse"
	leverIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/lever"
	smartRecruitersIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/smartrecruiters"
	workdayIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/workday"
	adzunaIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/jobSearch/adzuna"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type JobBoard interface {
	BoardType() string
	// ListJobs may skip re-reading jobs in knownExternalIDs and return them with DetailsSkipped.
	ListJobs(
		ctx context.Context,
		source *jobModel.JobSource,
		knownExternalIDs map[string]struct{},
	) (jobTypes.BoardListing, error)
}

type RegistryOptions struct {
	GreenhouseAPIURL      string
	LeverAPIURL           string
	LeverEUAPIURL         string
	AshbyAPIURL           string
	SmartRecruitersAPIURL string
	WorkdayHostURLFormat  string
	AdzunaAPIURL          string
	AdzunaAppID           string
	AdzunaAppKey          string
	Timeout               time.Duration
	// CareerPageTransport replaces the public-internet-only transport; only tests set it.
	CareerPageTransport http.RoundTripper
}

func NewJobBoardRegistry(options RegistryOptions) []JobBoard {
	if options.Timeout == 0 {
		options.Timeout = config.GetJobBoardTimeout()
	}

	return []JobBoard{
		&greenhouseBoard{
			integration: greenhouseIntegration.NewGreenhouseIntegration(
				options.GreenhouseAPIURL,
				options.Timeout,
			),
		},
		&leverBoard{
			global: leverIntegration.NewLeverIntegration(options.LeverAPIURL, options.Timeout),
			eu:     leverIntegration.NewLeverIntegration(options.LeverEUAPIURL, options.Timeout),
		},
		&ashbyBoard{
			integration: ashbyIntegration.NewAshbyIntegration(options.AshbyAPIURL, options.Timeout),
		},
		&smartRecruitersBoard{
			integration: smartRecruitersIntegration.NewSmartRecruitersIntegration(
				options.SmartRecruitersAPIURL,
				options.Timeout,
			),
		},
		&workdayBoard{
			integration: workdayIntegration.NewWorkdayIntegration(
				options.WorkdayHostURLFormat,
				options.Timeout,
			),
		},
		&adzunaBoard{
			integration: adzunaIntegration.NewAdzunaIntegration(
				options.AdzunaAPIURL,
				options.AdzunaAppID,
				options.AdzunaAppKey,
				options.Timeout,
			),
		},
		&careerPageBoard{
			fetcher: utils.NewWebFetcher(
				options.Timeout,
				config.GetCareerPageMaxBytes(),
				config.GetCareerPageUserAgent(),
				options.CareerPageTransport,
			),
		},
	}
}

func Find(registry []JobBoard, boardType string) JobBoard {
	for _, board := range registry {
		if board.BoardType() == boardType {
			return board
		}
	}

	return nil
}

// readNewJobs fetches the full record of postings not stored yet, up to the per-poll limit; postings
// already stored come back marked DetailsSkipped. A posting the board rejects (gone since it was
// listed) is dropped; other failures end the poll.
func readNewJobs(
	ctx context.Context,
	summaries []jobTypes.BoardJob,
	knownExternalIDs map[string]struct{},
	read func(ctx context.Context, summary jobTypes.BoardJob) (*jobTypes.BoardJob, error),
) ([]jobTypes.BoardJob, error) {
	jobs := make([]jobTypes.BoardJob, 0, len(summaries))
	readsLeft := config.GetJobBoardMaxDetailsPerPoll()

	for _, summary := range summaries {
		if _, known := knownExternalIDs[summary.ExternalID]; known {
			summary.DetailsSkipped = true
			jobs = append(jobs, summary)

			continue
		}

		if readsLeft == 0 {
			continue
		}

		readsLeft--

		job, err := read(ctx, summary)
		if err != nil && utils.ClassifyUpstreamError(err) == config.GetErrCodeUpstreamRejected() {
			continue
		}

		if err != nil {
			return nil, err
		}

		if job != nil {
			jobs = append(jobs, *job)
		}
	}

	return jobs, nil
}
