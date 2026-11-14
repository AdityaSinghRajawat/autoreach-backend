package jobManagement

import (
	"encoding/json"
	"strings"
	"time"

	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type IngestJobItem struct {
	Source         string     `json:"source"         validate:"required,oneof=LINKEDIN INDEED GLASSDOOR ZIPRECRUITER GREENHOUSE LEVER WORKDAY DIRECT"`
	CompanyName    string     `json:"companyName"    validate:"required,max=255"`
	CompanyDomain  string     `json:"companyDomain"  validate:"required,fqdn,max=255"`
	JobTitle       string     `json:"jobTitle"       validate:"required,max=255,validate_job_title"`
	JobDescription string     `json:"jobDescription" validate:"required,max=100000"`
	Location       *string    `json:"location"       validate:"omitnil,max=255"`
	Department     *string    `json:"department"     validate:"omitnil,max=255"`
	SalaryMin      *float64   `json:"salaryMin"      validate:"omitnil,gte=0"`
	SalaryMax      *float64   `json:"salaryMax"      validate:"omitnil,gte=0,validate_salary_range"`
	SalaryCurrency *string    `json:"salaryCurrency" validate:"omitnil,iso4217"`
	ATSType        string     `json:"atsType"        validate:"required,oneof=GREENHOUSE LEVER WORKDAY OTHER"`
	ExternalURL    string     `json:"externalUrl"    validate:"required,url,max=1000,validate_external_url"`
	ExternalJobID  *string    `json:"externalJobId"  validate:"omitnil,max=255"`
	PostedAt       *time.Time `json:"postedAt"`
}

type IngestJobsRequest struct {
	Jobs []IngestJobItem `json:"jobs" validate:"required,min=1,max=500,dive"`
}

type QueuedTask struct {
	JobID    string `json:"jobId"`
	TaskName string `json:"taskName"`
	Status   string `json:"status"`
}

type IngestJobsResponse struct {
	IngestedCount int          `json:"ingestedCount"`
	QueuedTasks   []QueuedTask `json:"queuedTasks"`
}

func (i *IngestJobItem) ToJobPostingModel() *jobModel.JobPosting {
	job := jobModel.NewJobPosting()
	job.Source = i.Source
	job.CompanyName = strings.TrimSpace(i.CompanyName)
	job.CompanyDomain = i.CompanyDomain
	job.JobTitle = strings.TrimSpace(i.JobTitle)
	job.JobDescription = strings.TrimSpace(i.JobDescription)
	job.Location = i.Location
	job.Department = i.Department
	job.SalaryMin = i.SalaryMin
	job.SalaryMax = i.SalaryMax
	job.SalaryCurrency = i.SalaryCurrency
	job.ATSType = i.ATSType
	job.ExternalURL = i.ExternalURL
	job.ExternalJobID = i.ExternalJobID
	job.PostedAt = i.PostedAt

	if rawPayload, err := json.Marshal(i); err == nil {
		job.RawPayload = rawPayload
	}

	job.ComputeHashes()

	return job
}

func (r *IngestJobsRequest) ToJobPostingModels() []*jobModel.JobPosting {
	jobs := make([]*jobModel.JobPosting, 0, len(r.Jobs))
	for index := range r.Jobs {
		jobs = append(jobs, r.Jobs[index].ToJobPostingModel())
	}

	return jobs
}

func NewIngestJobsResponse() *IngestJobsResponse {
	return &IngestJobsResponse{QueuedTasks: []QueuedTask{}}
}

func NewQueuedTask(jobID, taskName, status string) QueuedTask {
	return QueuedTask{JobID: jobID, TaskName: taskName, Status: status}
}
