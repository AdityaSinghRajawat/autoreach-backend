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
