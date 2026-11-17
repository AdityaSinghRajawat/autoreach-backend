package jobBoards

import (
	"context"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	workdayIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/workday"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type workdayBoard struct {
	integration *workdayIntegration.WorkdayIntegration
}

func (b *workdayBoard) BoardType() string { return string(modelHelpers.BoardTypeWorkday) }

// No official API: reads the JSON behind public career sites.
func (b *workdayBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	knownExternalIDs map[string]struct{},
) (jobTypes.BoardListing, error) {
	host, siteName, _ := strings.Cut(source.BoardToken, config.GetWorkdayTokenSeparator())
	tenant, instance, _ := strings.Cut(host, ".")
	site := workdayIntegration.Site{Tenant: tenant, Instance: instance, Name: siteName}

	summaries, complete, err := b.integration.ListJobs(ctx, site)
	if err != nil {
		return jobTypes.BoardListing{}, err
	}

	jobs, err := readNewJobs(
		ctx,
		summaries,
		knownExternalIDs,
		config.GetJobBoardDetailConcurrency(),
		func(ctx context.Context, summary jobTypes.BoardJob) (*jobTypes.BoardJob, error) {
			return b.integration.GetJob(ctx, site, summary)
		},
	)

	return jobTypes.BoardListing{Jobs: jobs, Complete: complete}, err
}
