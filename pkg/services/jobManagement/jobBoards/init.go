package jobBoards

import (
	"context"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	ashbyIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/ashby"
	greenhouseIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/greenhouse"
	leverIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/lever"
	smartRecruitersIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/smartrecruiters"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type JobBoard interface {
	ATSType() string
	// ListJobs may skip re-reading jobs in knownExternalIDs and return them with DetailsSkipped.
	ListJobs(
		ctx context.Context,
		source *jobModel.JobSource,
		knownExternalIDs map[string]struct{},
	) ([]jobTypes.BoardJob, error)
}

type RegistryOptions struct {
	GreenhouseAPIURL      string
	LeverAPIURL           string
	LeverEUAPIURL         string
	AshbyAPIURL           string
	SmartRecruitersAPIURL string
	Timeout               time.Duration
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
	}
}

func Find(registry []JobBoard, atsType string) JobBoard {
	for _, board := range registry {
		if board.ATSType() == atsType {
			return board
		}
	}

	return nil
}

type greenhouseBoard struct {
	integration *greenhouseIntegration.GreenhouseIntegration
}

func (b *greenhouseBoard) ATSType() string { return string(modelHelpers.ATSTypeGreenhouse) }

func (b *greenhouseBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) ([]jobTypes.BoardJob, error) {
	return b.integration.ListBoardJobs(ctx, source.BoardToken)
}

type leverBoard struct {
	global *leverIntegration.LeverIntegration
	eu     *leverIntegration.LeverIntegration
}

func (b *leverBoard) ATSType() string { return string(modelHelpers.ATSTypeLever) }

func (b *leverBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) ([]jobTypes.BoardJob, error) {
	if string(source.Region) == config.GetJobSourceRegionEU() {
		return b.eu.ListPostings(ctx, source.BoardToken)
	}

	return b.global.ListPostings(ctx, source.BoardToken)
}

type ashbyBoard struct {
	integration *ashbyIntegration.AshbyIntegration
}

func (b *ashbyBoard) ATSType() string { return string(modelHelpers.ATSTypeAshby) }

func (b *ashbyBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) ([]jobTypes.BoardJob, error) {
	return b.integration.ListJobs(ctx, source.BoardToken)
}

type smartRecruitersBoard struct {
	integration *smartRecruitersIntegration.SmartRecruitersIntegration
}

func (b *smartRecruitersBoard) ATSType() string { return string(modelHelpers.ATSTypeSmartRecruiters) }

// The listing has no descriptions, so each new posting costs one more request. Known postings are
// not re-read, and new ones beyond the per-poll limit wait for a later poll.
func (b *smartRecruitersBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	knownExternalIDs map[string]struct{},
) ([]jobTypes.BoardJob, error) {
	summaries, err := b.integration.ListPostings(ctx, source.BoardToken)
	if err != nil {
		return nil, err
	}

	jobs := make([]jobTypes.BoardJob, 0, len(summaries))
	detailsLeft := config.GetSmartRecruitersMaxDetailsPerPoll()

	for _, summary := range summaries {
		if _, known := knownExternalIDs[summary.ExternalID]; known {
			summary.DetailsSkipped = true
			jobs = append(jobs, summary)

			continue
		}

		if detailsLeft == 0 {
			continue
		}

		detailsLeft--

		job, err := b.integration.GetPosting(ctx, source.BoardToken, summary.ExternalID)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
