package jobBoards

import (
	"context"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	ashbyIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/ashby"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type ashbyBoard struct {
	integration *ashbyIntegration.AshbyIntegration
}

func (b *ashbyBoard) BoardType() string { return string(modelHelpers.BoardTypeAshby) }

func (b *ashbyBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) (jobTypes.BoardListing, error) {
	jobs, err := b.integration.ListJobs(ctx, source.BoardToken)

	return jobTypes.BoardListing{Jobs: jobs, Complete: true}, err
}
