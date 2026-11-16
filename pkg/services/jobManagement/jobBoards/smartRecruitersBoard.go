package jobBoards

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	smartRecruitersIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/smartrecruiters"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type smartRecruitersBoard struct {
	integration *smartRecruitersIntegration.SmartRecruitersIntegration
}

func (b *smartRecruitersBoard) BoardType() string {
	return string(modelHelpers.BoardTypeSmartRecruiters)
}

// The listing has no descriptions, so each new posting costs one more request.
func (b *smartRecruitersBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	knownExternalIDs map[string]struct{},
) (jobTypes.BoardListing, error) {
	summaries, err := b.integration.ListPostings(ctx, source.BoardToken)
	if err != nil {
		return jobTypes.BoardListing{}, err
	}

	jobs, err := readNewJobs(
		ctx,
		summaries,
		knownExternalIDs,
		config.GetJobBoardDetailConcurrency(),
		func(ctx context.Context, summary jobTypes.BoardJob) (*jobTypes.BoardJob, error) {
			job, err := b.integration.GetPosting(ctx, source.BoardToken, summary.ExternalID)

			return &job, err
		},
	)

	return jobTypes.BoardListing{Jobs: jobs, Complete: true}, err
}
