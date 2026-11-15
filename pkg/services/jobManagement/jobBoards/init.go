package jobBoards

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	greenhouseIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/greenhouse"
	leverIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/lever"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type JobBoard interface {
	ATSType() string
	ListJobs(ctx context.Context, source *jobModel.JobSource) ([]jobTypes.BoardJob, error)
}

type RegistryOptions struct {
	GreenhouseAPIURL string
	LeverAPIURL      string
	LeverEUAPIURL    string
}

type greenhouseBoard struct {
	integration *greenhouseIntegration.GreenhouseIntegration
}

type leverBoard struct {
	global *leverIntegration.LeverIntegration
	eu     *leverIntegration.LeverIntegration
}

func NewJobBoardRegistry(options RegistryOptions) []JobBoard {
	return []JobBoard{
		&greenhouseBoard{
			integration: greenhouseIntegration.NewGreenhouseIntegration(options.GreenhouseAPIURL),
		},
		&leverBoard{
			global: leverIntegration.NewLeverIntegration(options.LeverAPIURL),
			eu:     leverIntegration.NewLeverIntegration(options.LeverEUAPIURL),
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

func (b *greenhouseBoard) ATSType() string { return string(modelHelpers.ATSTypeGreenhouse) }

func (b *greenhouseBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
) ([]jobTypes.BoardJob, error) {
	return b.integration.ListBoardJobs(ctx, source.BoardToken)
}

func (b *leverBoard) ATSType() string { return string(modelHelpers.ATSTypeLever) }

func (b *leverBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
) ([]jobTypes.BoardJob, error) {
	if string(source.Region) == config.GetJobSourceRegionEU() {
		return b.eu.ListPostings(ctx, source.BoardToken)
	}

	return b.global.ListPostings(ctx, source.BoardToken)
}
