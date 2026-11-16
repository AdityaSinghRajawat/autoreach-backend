package workday

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (i *WorkdayIntegration) GetJob(
	ctx context.Context,
	site Site,
	summary jobTypes.BoardJob,
) (*jobTypes.BoardJob, error) {
	response := jobResponse{}
	path := fmt.Sprintf(
		config.GetWorkdayJobPathFormat(),
		site.Tenant,
		site.Name,
		summary.ExternalID,
	)

	if err := i.client(site).Get(ctx, path, nil, &response); err != nil {
		return nil, fmt.Errorf("get workday job: %w", err)
	}

	info := response.JobPostingInfo
	job := summary
	job.Title = utils.FirstNonEmpty(info.Title, summary.Title)
	job.Description = utils.HTMLToText(info.JobDescription)
	job.URL = info.ExternalURL

	if info.Location != "" {
		job.Location = &info.Location
	}

	if startDate, err := time.Parse(config.GetWorkdayDateLayout(), info.StartDate); err == nil {
		job.PostedAt = &startDate
	}

	return &job, nil
}
