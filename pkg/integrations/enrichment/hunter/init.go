package hunter

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	SearchResponse struct {
		Data struct {
			Emails []Email `json:"emails"`
		} `json:"data"`
	}

	Email struct {
		Value        string        `json:"value"`
		FirstName    string        `json:"first_name"`
		LastName     string        `json:"last_name"`
		Position     string        `json:"position"`
		LinkedIn     string        `json:"linkedin"`
		Verification *Verification `json:"verification"`
	}

	Verification struct {
		Status string `json:"status"`
	}
)

type HunterIntegration struct {
	httpClient *utils.HTTPClient
	apiKey     string
}

func NewHunterIntegration(policy config.ProviderPolicy) *HunterIntegration {
	return &HunterIntegration{
		httpClient: utils.NewHTTPClient(policy.BaseURL, policy.Timeout),
		apiKey:     policy.APIKey,
	}
}
