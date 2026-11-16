package adzuna

import (
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	searchResponse struct {
		Results []result `json:"results"`
	}

	result struct {
		ID                string     `json:"id"`
		Title             string     `json:"title"`
		Description       string     `json:"description"`
		RedirectURL       string     `json:"redirect_url"`
		Created           *time.Time `json:"created"`
		SalaryMin         *float64   `json:"salary_min"`
		SalaryMax         *float64   `json:"salary_max"`
		SalaryIsPredicted string     `json:"salary_is_predicted"`
		Company           struct {
			DisplayName string `json:"display_name"`
		} `json:"company"`
		Location struct {
			DisplayName string `json:"display_name"`
		} `json:"location"`
		Category struct {
			Label string `json:"label"`
		} `json:"category"`
	}
)

type AdzunaIntegration struct {
	httpClient *utils.HttpClient
	appID      string
	appKey     string
}

func NewAdzunaIntegration(baseURL, appID, appKey string, timeout time.Duration) *AdzunaIntegration {
	return &AdzunaIntegration{
		httpClient: utils.NewHttpClient(baseURL, timeout),
		appID:      appID,
		appKey:     appKey,
	}
}
