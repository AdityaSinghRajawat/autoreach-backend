package workday

import (
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	listRequest struct {
		AppliedFacets map[string]any `json:"appliedFacets"`
		Limit         int            `json:"limit"`
		Offset        int            `json:"offset"`
		SearchText    string         `json:"searchText"`
	}

	listResponse struct {
		Total       int          `json:"total"`
		JobPostings []jobSummary `json:"jobPostings"`
	}

	jobSummary struct {
		Title         string `json:"title"`
		ExternalPath  string `json:"externalPath"`
		LocationsText string `json:"locationsText"`
	}

	jobResponse struct {
		JobPostingInfo struct {
			Title          string `json:"title"`
			JobDescription string `json:"jobDescription"`
			Location       string `json:"location"`
			StartDate      string `json:"startDate"`
			ExternalURL    string `json:"externalUrl"`
		} `json:"jobPostingInfo"`
	}
)

// Site identifies one Workday career site, as in https://{tenant}.{instance}.myworkdayjobs.com/{name}.
type Site struct {
	Tenant   string
	Instance string
	Name     string
}

type WorkdayIntegration struct {
	hostURLFormat string
	timeout       time.Duration
}

func NewWorkdayIntegration(hostURLFormat string, timeout time.Duration) *WorkdayIntegration {
	return &WorkdayIntegration{hostURLFormat: hostURLFormat, timeout: timeout}
}

func (i *WorkdayIntegration) client(site Site) *utils.HTTPClient {
	return utils.NewHTTPClient(fmt.Sprintf(i.hostURLFormat, site.Tenant, site.Instance), i.timeout)
}
