package jobBoards

import (
	"context"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	greenhouseIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/greenhouse"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type greenhouseBoard struct {
	integration *greenhouseIntegration.GreenhouseIntegration
}

func (b *greenhouseBoard) BoardType() string { return string(modelHelpers.BoardTypeGreenhouse) }

func (b *greenhouseBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) (jobTypes.BoardListing, error) {
	jobs, err := b.integration.ListBoardJobs(ctx, source.BoardToken)

	return jobTypes.BoardListing{Jobs: jobs, Complete: true}, err
}
