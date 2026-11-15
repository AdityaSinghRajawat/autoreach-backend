package jobManagement

import (
	"strings"
	"time"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type CreateJobSourceRequest struct {
	ATSType       string  `json:"atsType"       validate:"required,oneof=GREENHOUSE LEVER"`
	BoardToken    string  `json:"boardToken"    validate:"required,max=100"`
	Region        *string `json:"region"        validate:"omitnil,oneof=GLOBAL EU"`
	CompanyName   string  `json:"companyName"   validate:"required,max=255"`
	CompanyDomain string  `json:"companyDomain" validate:"required,fqdn,max=255"`
}

type UpdateJobSourceRequest struct {
	CompanyName   *string `json:"companyName"   validate:"omitnil,min=1,max=255"`
	CompanyDomain *string `json:"companyDomain" validate:"omitnil,fqdn,max=255"`
	IsActive      *bool   `json:"isActive"`
}

type PollJobSourceResult struct {
	SourceID string `json:"sourceId"`
	Fetched  int    `json:"fetched"`
	Ingested int    `json:"ingested"`
	Skipped  int    `json:"skipped"`
	Closed   int    `json:"closed"`
}

type BoardJob struct {
	ExternalID  string
	Title       string
	Description string
	Location    *string
	Department  *string
	URL         string
	PostedAt    *time.Time
}

func (r *CreateJobSourceRequest) ToJobSourceModel() *jobModel.JobSource {
	source := jobModel.NewJobSource()
	source.ATSType = r.ATSType
	source.BoardToken = strings.ToLower(strings.TrimSpace(r.BoardToken))
	source.CompanyName = strings.TrimSpace(r.CompanyName)
	source.CompanyDomain = strings.ToLower(r.CompanyDomain)

	if r.Region != nil {
		source.Region = modelHelpers.JobSourceRegion(*r.Region)
	}

	return source
}

func (r *UpdateJobSourceRequest) IsEmpty() bool {
	return r.CompanyName == nil && r.CompanyDomain == nil && r.IsActive == nil
}

func (r *UpdateJobSourceRequest) ToJobSourceModel(
	existing *jobModel.JobSource,
) *jobModel.JobSource {
	source := *existing

	if r.CompanyName != nil {
		source.CompanyName = strings.TrimSpace(*r.CompanyName)
	}

	if r.CompanyDomain != nil {
		source.CompanyDomain = strings.ToLower(*r.CompanyDomain)
	}

	if r.IsActive != nil {
		source.IsActive = *r.IsActive
	}

	return &source
}

func (j BoardJob) ToIngestJobItem(source *jobModel.JobSource) IngestJobItem {
	externalID := j.ExternalID

	return IngestJobItem{
		Source:         source.ATSType,
		CompanyName:    source.CompanyName,
		CompanyDomain:  source.CompanyDomain,
		JobTitle:       j.Title,
		JobDescription: j.Description,
		Location:       j.Location,
		Department:     j.Department,
		ATSType:        source.ATSType,
		ExternalURL:    j.URL,
		ExternalJobID:  &externalID,
		PostedAt:       j.PostedAt,
	}
}
