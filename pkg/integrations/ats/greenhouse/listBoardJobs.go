package greenhouse

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (i *GreenhouseIntegration) ListBoardJobs(
	ctx context.Context,
	boardToken string,
) ([]jobTypes.BoardJob, error) {
	response := boardJobsResponse{}

	err := i.httpClient.Get(
		ctx,
		"/boards/"+boardToken+"/jobs"+config.GetGreenhouseBoardJobsQuery(),
		nil,
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("list greenhouse board jobs: %w", err)
	}

	jobs := make([]jobTypes.BoardJob, 0, len(response.Jobs))

	for _, item := range response.Jobs {
		job := jobTypes.BoardJob{
			ExternalID:  strconv.FormatInt(item.ID, 10),
			Title:       item.Title,
			Description: utils.HTMLToText(item.Content),
			URL:         fmt.Sprintf(config.GetGreenhouseJobURLFormat(), boardToken, item.ID),
			PostedAt:    item.UpdatedAt,
			Location:    utils.NilIfEmpty(item.Location.Name),
		}

		if len(item.Departments) > 0 {
			job.Department = utils.NilIfEmpty(item.Departments[0].Name)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
