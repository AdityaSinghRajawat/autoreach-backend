package smartrecruiters

import (
	"strings"
	"time"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	postingsPage struct {
		Offset     int       `json:"offset"`
		TotalFound int       `json:"totalFound"`
		Content    []posting `json:"content"`
	}

	posting struct {
		ID           string     `json:"id"`
		Name         string     `json:"name"`
		ReleasedDate *time.Time `json:"releasedDate"`
		PostingURL   string     `json:"postingUrl"`
		Location     struct {
			City         string `json:"city"`
			FullLocation string `json:"fullLocation"`
		} `json:"location"`
		Department struct {
			Label string `json:"label"`
		} `json:"department"`
		JobAd struct {
			Sections struct {
				CompanyDescription    section `json:"companyDescription"`
				JobDescription        section `json:"jobDescription"`
				Qualifications        section `json:"qualifications"`
				AdditionalInformation section `json:"additionalInformation"`
			} `json:"sections"`
		} `json:"jobAd"`
	}

	section struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}
)

type SmartRecruitersIntegration struct {
	httpClient *utils.HTTPClient
}

func NewSmartRecruitersIntegration(
	baseURL string,
	timeout time.Duration,
) *SmartRecruitersIntegration {
	return &SmartRecruitersIntegration{httpClient: utils.NewHTTPClient(baseURL, timeout)}
}

func (p posting) toBoardJob() jobTypes.BoardJob {
	sections := p.JobAd.Sections
	parts := []string{}

	for _, part := range []section{
		sections.CompanyDescription,
		sections.JobDescription,
		sections.Qualifications,
		sections.AdditionalInformation,
	} {
		if text := utils.HTMLToText(part.Text); text != "" {
			parts = append(parts, strings.TrimSpace(part.Title+"\n"+text))
		}
	}

	return jobTypes.BoardJob{
		ExternalID:  p.ID,
		Title:       p.Name,
		Description: strings.Join(parts, "\n\n"),
		URL:         p.PostingURL,
		PostedAt:    p.ReleasedDate,
		Location: utils.NilIfEmpty(
			utils.FirstNonEmpty(p.Location.FullLocation, p.Location.City),
		),
		Department: utils.NilIfEmpty(p.Department.Label),
	}
}
