package lever

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
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
	httpClient *utils.HttpClient
}

func NewLeverIntegration(baseURL string) *LeverIntegration {
	return &LeverIntegration{httpClient: utils.NewHttpClient(baseURL, config.GetATSTimeout())}
}

func postingPath(site, postingID string) string {
	return "/postings/" + site + "/" + postingID
}
