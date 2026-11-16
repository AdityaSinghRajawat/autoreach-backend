package ashby

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (i *AshbyIntegration) ListJobs(
	ctx context.Context,
	jobBoardName string,
) ([]jobTypes.BoardJob, error) {
	response := jobBoardResponse{}

	if err := i.httpClient.Get(
		ctx,
		config.GetAshbyJobBoardPath()+jobBoardName,
		nil,
		&response,
	); err != nil {
		return nil, fmt.Errorf("list ashby jobs: %w", err)
	}

	jobs := make([]jobTypes.BoardJob, 0, len(response.Jobs))

	for _, item := range response.Jobs {
		if !item.IsListed {
			continue
		}

		description := strings.TrimSpace(item.DescriptionPlain)
		if description == "" {
			description = utils.HTMLToText(item.DescriptionHTML)
		}

		jobs = append(jobs, jobTypes.BoardJob{
			ExternalID:  item.ID,
			Title:       item.Title,
			Description: description,
			URL:         item.JobURL,
			PostedAt:    item.PublishedAt,
			Location:    utils.NilIfEmpty(item.Location),
			Department:  utils.NilIfEmpty(utils.FirstNonEmpty(item.Department, item.Team)),
		})
	}

	return jobs, nil
}
