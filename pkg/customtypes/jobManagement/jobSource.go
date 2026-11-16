package jobManagement

import (
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type CreateJobSourceRequest struct {
	BoardType      string  `json:"boardType"      validate:"required,oneof=GREENHOUSE LEVER ASHBY SMARTRECRUITERS WORKDAY ADZUNA CAREER_PAGE"`
	BoardToken     string  `json:"boardToken"     validate:"required,max=500"`
	SearchQuery    *string `json:"searchQuery"    validate:"omitnil,max=255"`
	SearchLocation *string `json:"searchLocation" validate:"omitnil,max=255"`
	Region         *string `json:"region"         validate:"omitnil,oneof=GLOBAL EU"`
	CompanyName    *string `json:"companyName"    validate:"omitnil,max=255"`
	CompanyDomain  *string `json:"companyDomain"  validate:"omitnil,max=255"`
}

type UpdateJobSourceRequest struct {
	SearchQuery    *string `json:"searchQuery"    validate:"omitnil,min=1,max=255"`
	SearchLocation *string `json:"searchLocation" validate:"omitnil,min=1,max=255"`
	CompanyName    *string `json:"companyName"    validate:"omitnil,min=1,max=255"`
	CompanyDomain  *string `json:"companyDomain"  validate:"omitnil,fqdn,max=255"`
	IsActive       *bool   `json:"isActive"`
}

type PollJobSourceResult struct {
	SourceID  string `json:"sourceId"`
	Fetched   int    `json:"fetched"`
	Ingested  int    `json:"ingested"`
	Skipped   int    `json:"skipped"`
	Closed    int    `json:"closed"`
	Unchanged int    `json:"unchanged"`
}

type BoardJob struct {
	ExternalID     string
	Title          string
	Description    string
	Location       *string
	Department     *string
	URL            string
	PostedAt       *time.Time
	SalaryMin      *float64
	SalaryMax      *float64
	SalaryCurrency *string
	// CompanyName and CompanyDomain are set only by boards that list many companies.
	CompanyName   string
	CompanyDomain string
	// DetailsSkipped marks a job already stored that the board did not re-read.
	DetailsSkipped bool
}

// BoardListing is one poll's jobs; Complete is false when the board may not list every open job.
type BoardListing struct {
	Jobs     []BoardJob
	Complete bool
}

func (r *CreateJobSourceRequest) ToJobSourceModel() *jobModel.JobSource {
	source := jobModel.NewJobSource()
	source.BoardType = modelHelpers.BoardType(r.BoardType)
	source.BoardToken = strings.TrimSpace(r.BoardToken)
	source.SearchQuery = trimmed(r.SearchQuery)
	source.SearchLocation = trimmed(r.SearchLocation)
	source.CompanyName = trimmed(r.CompanyName)
	source.CompanyDomain = lowered(r.CompanyDomain)

	if r.BoardType != config.GetCareerPageBoardType() {
		source.BoardToken = strings.ToLower(source.BoardToken)
	}

	if r.Region != nil {
		source.Region = modelHelpers.JobSourceRegion(*r.Region)
	}

	return source
}

func (r *UpdateJobSourceRequest) IsEmpty() bool {
	return r.SearchQuery == nil && r.SearchLocation == nil && r.CompanyName == nil &&
		r.CompanyDomain == nil && r.IsActive == nil
}

func (r *UpdateJobSourceRequest) ToJobSourceModel(
	existing *jobModel.JobSource,
) *jobModel.JobSource {
	source := *existing

	if r.SearchQuery != nil {
		source.SearchQuery = trimmed(r.SearchQuery)
	}

	if r.SearchLocation != nil {
		source.SearchLocation = trimmed(r.SearchLocation)
	}

	if r.CompanyName != nil {
		source.CompanyName = trimmed(r.CompanyName)
	}

	if r.CompanyDomain != nil {
		source.CompanyDomain = lowered(r.CompanyDomain)
	}

	if r.IsActive != nil {
		source.IsActive = *r.IsActive
	}

	return &source
}

func (j BoardJob) ToIngestJobItem(source *jobModel.JobSource) IngestJobItem {
	jobSource, atsType := config.GetBoardJobKinds(string(source.BoardType))
	externalID := j.ExternalID
	companyName, companyDomain := j.CompanyName, j.CompanyDomain

	if companyName == "" && source.CompanyName != nil {
		companyName = *source.CompanyName
	}

	if companyDomain == "" && source.CompanyDomain != nil {
		companyDomain = *source.CompanyDomain
	}

	return IngestJobItem{
		Source:         jobSource,
		CompanyName:    companyName,
		CompanyDomain:  companyDomain,
		JobTitle:       j.Title,
		JobDescription: j.Description,
		Location:       j.Location,
		Department:     j.Department,
		SalaryMin:      j.SalaryMin,
		SalaryMax:      j.SalaryMax,
		SalaryCurrency: j.SalaryCurrency,
		ATSType:        atsType,
		ExternalURL:    j.URL,
		ExternalJobID:  &externalID,
		PostedAt:       j.PostedAt,
	}
}

func trimmed(value *string) *string {
	if value == nil {
		return nil
	}

	text := strings.TrimSpace(*value)

	return &text
}

func lowered(value *string) *string {
	if value == nil {
		return nil
	}

	text := strings.ToLower(strings.TrimSpace(*value))

	return &text
}
