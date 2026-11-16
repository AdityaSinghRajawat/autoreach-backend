package ashby

import (
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	jobBoardResponse struct {
		Jobs []job `json:"jobs"`
	}

	job struct {
		ID               string     `json:"id"`
		Title            string     `json:"title"`
		Department       string     `json:"department"`
		Team             string     `json:"team"`
		Location         string     `json:"location"`
		IsListed         bool       `json:"isListed"`
		PublishedAt      *time.Time `json:"publishedAt"`
		JobURL           string     `json:"jobUrl"`
		DescriptionPlain string     `json:"descriptionPlain"`
		DescriptionHTML  string     `json:"descriptionHtml"`
	}
)

type AshbyIntegration struct {
	httpClient *utils.HttpClient
}

func NewAshbyIntegration(baseURL string, timeout time.Duration) *AshbyIntegration {
	return &AshbyIntegration{httpClient: utils.NewHttpClient(baseURL, timeout)}
}
