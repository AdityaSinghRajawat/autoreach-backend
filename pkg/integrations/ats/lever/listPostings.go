package lever

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (i *LeverIntegration) ListPostings(
	ctx context.Context,
	site string,
) ([]jobTypes.BoardJob, error) {
	postings := []posting{}

	if err := i.httpClient.Get(
		ctx,
		"/postings/"+site+config.GetLeverPostingsQuery(),
		nil,
		&postings,
	); err != nil {
		return nil, fmt.Errorf("list lever postings: %w", err)
	}

	jobs := make([]jobTypes.BoardJob, 0, len(postings))

	for _, item := range postings {
		sections := []string{item.DescriptionPlain}
		for _, list := range item.Lists {
			sections = append(sections, list.Text+"\n"+utils.HTMLToText(list.Content))
		}

		sections = append(sections, item.AdditionalPlain)

		job := jobTypes.BoardJob{
			ExternalID:  item.ID,
			Title:       item.Text,
			Description: strings.TrimSpace(strings.Join(sections, "\n\n")),
			URL:         item.HostedURL,
			Location:    optional(item.Categories.Location),
			Department:  optional(firstNonEmpty(item.Categories.Department, item.Categories.Team)),
		}

		if item.CreatedAt > 0 {
			postedAt := time.UnixMilli(item.CreatedAt).UTC()
			job.PostedAt = &postedAt
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func optional(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
