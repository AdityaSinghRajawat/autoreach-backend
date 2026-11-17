package signalhire

import (
	"encoding/json"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	searchRequest struct {
		Items       []string `json:"items"`
		CallbackURL string   `json:"callbackUrl"`
	}

	SearchResponse struct {
		RequestID json.RawMessage `json:"requestId"`
	}

	searchByQueryRequest struct {
		CurrentCompany string `json:"currentCompany"`
		CurrentTitle   string `json:"currentTitle,omitempty"`
		Size           int    `json:"size"`
	}

	SearchByQueryResponse struct {
		Profiles []Profile `json:"profiles"`
	}

	Profile struct {
		UID        string       `json:"uid"`
		FullName   string       `json:"fullName"`
		Experience []Experience `json:"experience"`
	}

	Experience struct {
		Company string `json:"company"`
		Title   string `json:"title"`
	}

	CallbackItem struct {
		Item      string     `json:"item"`
		Status    string     `json:"status"`
		Candidate *Candidate `json:"candidate"`
	}

	Candidate struct {
		FullName string           `json:"fullName"`
		Headline string           `json:"headLine"`
		Contacts []ContactChannel `json:"contacts"`
	}

	ContactChannel struct {
		Type    string `json:"type"`
		Value   string `json:"value"`
		Rating  string `json:"rating"`
		SubType string `json:"subType"`
	}
)

type SignalHireIntegration struct {
	httpClient *utils.HTTPClient
	apiKey     string
}

func NewSignalHireIntegration(policy config.ProviderPolicy) *SignalHireIntegration {
	return &SignalHireIntegration{
		httpClient: utils.NewHTTPClient(policy.BaseURL, policy.Timeout),
		apiKey:     policy.APIKey,
	}
}
