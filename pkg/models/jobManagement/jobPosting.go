package jobManagement

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type JobPosting struct {
	ID              string                 `db:"id"               json:"jobId"          validate:"required,uuid"`
	JobHash         string                 `db:"job_hash"         json:"-"              validate:"required,len=64,hexadecimal"`
	Source          string                 `db:"source"           json:"source"         validate:"required,oneof=LINKEDIN INDEED GLASSDOOR ZIPRECRUITER NAUKRI WELLFOUND GREENHOUSE LEVER WORKDAY ASHBY SMARTRECRUITERS ADZUNA DIRECT"`
	CompanyName     string                 `db:"company_name"     json:"companyName"    validate:"required,max=255"`
	CompanyDomain   string                 `db:"company_domain"   json:"companyDomain"  validate:"required,fqdn,max=255"`
	JobTitle        string                 `db:"job_title"        json:"jobTitle"       validate:"required,max=255,validate_job_title"`
	NormalizedTitle string                 `db:"normalized_title" json:"-"              validate:"required,max=255"`
	JobDescription  string                 `db:"job_description"  json:"jobDescription" validate:"required,max=100000"`
	DescriptionHash string                 `db:"description_hash" json:"-"              validate:"required,len=64,hexadecimal"`
	Location        *string                `db:"location"         json:"location"       validate:"omitnil,max=255"`
	Department      *string                `db:"department"       json:"department"     validate:"omitnil,max=255"`
	SalaryMin       *float64               `db:"salary_min"       json:"salaryMin"      validate:"omitnil,gte=0"`
	SalaryMax       *float64               `db:"salary_max"       json:"salaryMax"      validate:"omitnil,gte=0,validate_salary_range"`
	SalaryCurrency  *string                `db:"salary_currency"  json:"salaryCurrency" validate:"omitnil,iso4217"`
	ATSType         string                 `db:"ats_type"         json:"atsType"        validate:"required,oneof=GREENHOUSE LEVER WORKDAY ASHBY SMARTRECRUITERS OTHER"`
	ExternalURL     string                 `db:"external_url"     json:"externalUrl"    validate:"required,max=1000,validate_external_url"`
	JobSourceID     *string                `db:"job_source_id"    json:"-"              validate:"omitnil,uuid"`
	ExternalJobID   *string                `db:"external_job_id"  json:"externalJobId"  validate:"omitnil,max=255"`
	RawPayload      json.RawMessage        `db:"raw_payload"      json:"-"              validate:"required,validate_json_object"`
	Status          modelHelpers.JobStatus `db:"status"           json:"status"         validate:"required,oneof=ACTIVE CLOSED ARCHIVED"`
	IsEmbedded      bool                   `db:"is_embedded"      json:"isEmbedded"`
	PostedAt        *time.Time             `db:"posted_at"        json:"postedAt"`
	LastSeenAt      time.Time              `db:"last_seen_at"     json:"lastSeenAt"     validate:"required"`
	CreatedAt       time.Time              `db:"created_at"       json:"createdAt"      validate:"required"`
	UpdatedAt       time.Time              `db:"updated_at"       json:"updatedAt"      validate:"required"`
}

func NewJobPosting() *JobPosting {
	now := utils.GetCurrTimeStamp()

	return &JobPosting{
		ID:         utils.GenerateUUID(),
		RawPayload: json.RawMessage(config.GetEmptyJSONObject()),
		Status:     modelHelpers.JobStatusActive,
		LastSeenAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (j *JobPosting) UpdateTimestamp() {
	now := utils.GetCurrTimeStamp()
	j.LastSeenAt = now
	j.UpdatedAt = now
}

func (j *JobPosting) Validate() error {
	if err := validations.Validate(j); err != nil {
		return fmt.Errorf("job posting validation failed: %w", err)
	}

	return nil
}

func (j *JobPosting) ComputeHashes() {
	j.CompanyDomain = modelHelpers.NormalizeDomain(j.CompanyDomain)
	j.NormalizedTitle = modelHelpers.NormalizeJobTitle(j.JobTitle)
	j.JobHash = ComputeJobHash(j.CompanyDomain, j.NormalizedTitle)
	j.DescriptionHash = utils.Fingerprint(j.JobDescription)
}

func ComputeJobHash(companyDomain, normalizedTitle string) string {
	return utils.Fingerprint(
		config.GetJobHashVersion() + "|" +
			modelHelpers.NormalizeDomain(companyDomain) + "|" +
			normalizedTitle,
	)
}
