package jobBoards

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	leverIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/lever"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type leverBoard struct {
	global *leverIntegration.LeverIntegration
	eu     *leverIntegration.LeverIntegration
}

func (b *leverBoard) BoardType() string { return string(modelHelpers.BoardTypeLever) }

func (b *leverBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) (jobTypes.BoardListing, error) {
	integration := b.global
	if string(source.Region) == config.GetJobSourceRegionEU() {
		integration = b.eu
	}

	jobs, err := integration.ListPostings(ctx, source.BoardToken)

	return jobTypes.BoardListing{Jobs: jobs, Complete: true}, err
}
