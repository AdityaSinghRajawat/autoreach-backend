package workday

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// ListJobs returns the site's postings without descriptions. Workday reports the total only on the
// first page and lists at most 2000 postings, so complete is false when that cap is reached.
func (i *WorkdayIntegration) ListJobs(
	ctx context.Context,
	site Site,
) ([]jobTypes.BoardJob, bool, error) {
	client := i.client(site)
	path := fmt.Sprintf(config.GetWorkdayJobsPathFormat(), site.Tenant, site.Name)
	jobs, total := []jobTypes.BoardJob{}, -1

	for offset := 0; offset < config.GetWorkdayMaxListed(); offset += config.GetWorkdayPageSize() {
		page := listResponse{}

		err := client.Post(ctx, path, nil, listRequest{
			AppliedFacets: map[string]any{},
			Limit:         config.GetWorkdayPageSize(),
			Offset:        offset,
		}, &page)
		if err != nil {
			return nil, false, fmt.Errorf("list workday jobs: %w", err)
		}

		if total < 0 {
			total = page.Total
		}

		for _, summary := range page.JobPostings {
			jobs = append(jobs, jobTypes.BoardJob{
				ExternalID: summary.ExternalPath,
				Title:      summary.Title,
				Location:   utils.NilIfEmpty(summary.LocationsText),
			})
		}

		if len(page.JobPostings) == 0 || offset+len(page.JobPostings) >= total {
			break
		}
	}

	return jobs, total < config.GetWorkdayMaxListed(), nil
}
