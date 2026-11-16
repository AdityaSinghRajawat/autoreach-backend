package jobBoards

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	adzunaIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/jobSearch/adzuna"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type adzunaBoard struct {
	integration *adzunaIntegration.AdzunaIntegration
}

func (b *adzunaBoard) BoardType() string { return string(modelHelpers.BoardTypeAdzuna) }

// A search returns only the best matches, so missing jobs are not closed on the next poll.
func (b *adzunaBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	_ map[string]struct{},
) (jobTypes.BoardListing, error) {
	where := ""
	if source.SearchLocation != nil {
		where = *source.SearchLocation
	}

	jobs := []jobTypes.BoardJob{}

	for page := 1; page <= config.GetAdzunaMaxPages(); page++ {
		results, err := b.integration.SearchJobs(
			ctx,
			source.BoardToken,
			*source.SearchQuery,
			where,
			page,
		)
		if err != nil {
			return jobTypes.BoardListing{}, err
		}

		jobs = append(jobs, results...)

		if len(results) < config.GetAdzunaResultsPerPage() {
			break
		}
	}

	return jobTypes.BoardListing{Jobs: jobs, Complete: false}, nil
}
