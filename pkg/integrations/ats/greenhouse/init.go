package greenhouse

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	jobResponse struct {
		Questions         []question `json:"questions"`
		LocationQuestions []question `json:"location_questions"`
	}

	question struct {
		Label    string  `json:"label"`
		Required bool    `json:"required"`
		Fields   []field `json:"fields"`
	}

	field struct {
		Name   string       `json:"name"`
		Type   string       `json:"type"`
		Values []fieldValue `json:"values"`
	}

	boardJobsResponse struct {
		Jobs []boardJob `json:"jobs"`
	}

	boardJob struct {
		ID          int64      `json:"id"`
		Title       string     `json:"title"`
		AbsoluteURL string     `json:"absolute_url"`
		Content     string     `json:"content"`
		UpdatedAt   *time.Time `json:"updated_at"`
		Location    struct {
			Name string `json:"name"`
		} `json:"location"`
		Departments []struct {
			Name string `json:"name"`
		} `json:"departments"`
	}

	fieldValue struct {
		Label string          `json:"label"`
		Value json.RawMessage `json:"value"`
	}
)

type GreenhouseIntegration struct {
	httpClient *utils.HttpClient
}

func NewGreenhouseIntegration(baseURL string) *GreenhouseIntegration {
	return &GreenhouseIntegration{httpClient: utils.NewHttpClient(baseURL, config.GetATSTimeout())}
}

func jobPath(boardToken, jobID string) string {
	return "/boards/" + boardToken + "/jobs/" + jobID
}

func basicAuth(apiKey string) map[string]string {
	return map[string]string{
		config.GetAuthorizationHeader(): "Basic " + base64.StdEncoding.EncodeToString(
			[]byte(apiKey+":"),
		),
	}
}

func rawValue(value json.RawMessage) string {
	var text string
	if json.Unmarshal(value, &text) == nil {
		return text
	}

	return strings.TrimSpace(string(value))
}
