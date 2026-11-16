package lever

import (
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	postingResponse struct {
		ID       string `json:"id"`
		Text     string `json:"text"`
		ApplyURL string `json:"applyUrl"`
	}

	posting struct {
		ID               string `json:"id"`
		Text             string `json:"text"`
		HostedURL        string `json:"hostedUrl"`
		DescriptionPlain string `json:"descriptionPlain"`
		AdditionalPlain  string `json:"additionalPlain"`
		CreatedAt        int64  `json:"createdAt"`
		Lists            []struct {
			Text    string `json:"text"`
			Content string `json:"content"`
		} `json:"lists"`
		Categories struct {
			Location   string `json:"location"`
			Team       string `json:"team"`
			Department string `json:"department"`
		} `json:"categories"`
	}

	applyResponse struct {
		OK            bool   `json:"ok"`
		ApplicationID string `json:"applicationId"`
	}
)

type LeverIntegration struct {
	httpClient *utils.HTTPClient
}

func NewLeverIntegration(baseURL string, timeout time.Duration) *LeverIntegration {
	return &LeverIntegration{httpClient: utils.NewHTTPClient(baseURL, timeout)}
}

func postingPath(site, postingID string) string {
	return "/postings/" + site + "/" + postingID
}
