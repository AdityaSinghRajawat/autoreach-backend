package adzuna

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// SearchJobs returns one page of results. Adzuna's descriptions are short snippets of the ad.
func (i *AdzunaIntegration) SearchJobs(
	ctx context.Context,
	country, what, where string,
	page int,
) ([]jobTypes.BoardJob, error) {
	if i.appID == "" || i.appKey == "" {
		return nil, fmt.Errorf("search adzuna jobs: %w", config.ErrMissingCredential)
	}

	query := url.Values{
		"app_id":           {i.appID},
		"app_key":          {i.appKey},
		"results_per_page": {strconv.Itoa(config.GetAdzunaResultsPerPage())},
		"what":             {what},
	}
	if where != "" {
		query.Set("where", where)
	}

	response := searchResponse{}
	path := fmt.Sprintf(config.GetAdzunaSearchPathFormat(), country, page) + "?" + query.Encode()

	if err := i.httpClient.Get(ctx, path, nil, &response); err != nil {
		return nil, fmt.Errorf("search adzuna jobs: %w", err)
	}

	currency, knownCurrency := config.GetAdzunaCurrency(country)
	jobs := make([]jobTypes.BoardJob, 0, len(response.Results))

	for _, item := range response.Results {
		job := jobTypes.BoardJob{
			ExternalID:  item.ID.String(),
			Title:       utils.HTMLToText(item.Title),
			Description: utils.HTMLToText(item.Description),
			URL:         item.RedirectURL,
			PostedAt:    item.Created,
			Location:    utils.NilIfEmpty(item.Location.DisplayName),
			Department:  utils.NilIfEmpty(item.Category.Label),
			CompanyName: item.Company.DisplayName,
		}

		if knownCurrency && item.SalaryIsPredicted.String() != config.GetAdzunaPredictedSalary() &&
			(item.SalaryMin != nil || item.SalaryMax != nil) {
			job.SalaryMin, job.SalaryMax, job.SalaryCurrency = item.SalaryMin, item.SalaryMax, &currency
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
