package applicationManagement

import (
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type CreateApplicationRequest struct {
	JobID string `json:"jobId" validate:"required,uuid"`
}

type UpdateApplicationRequest struct {
	Status string `json:"status" validate:"required,oneof=SUBMITTED INTERVIEWING REJECTED"`
}

type ListApplicationsParams struct {
	sharedTypes.PageParams
	Status string `form:"status" validate:"omitempty,oneof=QUEUED IN_PROGRESS SUBMITTED ACTION_REQUIRED FAILED INTERVIEWING REJECTED"`
}

type ApplicationJob struct {
	JobID       string  `json:"jobId"`
	JobTitle    string  `json:"jobTitle"`
	CompanyName string  `json:"companyName"`
	Location    *string `json:"location"`
}

type ApplicationDetail struct {
	*applicationModel.Application
	Job *ApplicationJob `json:"job"`
}

func (r *CreateApplicationRequest) ToApplicationModel(
	candidateID string,
	job *jobModel.JobPosting,
	matchScore float64,
) *applicationModel.Application {
	application := applicationModel.NewApplication()
	application.CandidateID = candidateID
	application.JobID = job.ID
	application.MatchScore = matchScore
	application.ATSType = job.ATSType
	application.ApplyURL = job.ExternalURL

	return application
}

func NewApplicationDetails(
	applications []*applicationModel.Application,
	jobs []*jobModel.JobPosting,
) []*ApplicationDetail {
	byID := make(map[string]*ApplicationJob, len(jobs))
	for _, job := range jobs {
		byID[job.ID] = &ApplicationJob{
			JobID:       job.ID,
			JobTitle:    job.JobTitle,
			CompanyName: job.CompanyName,
			Location:    job.Location,
		}
	}

	details := make([]*ApplicationDetail, 0, len(applications))
	for _, application := range applications {
		details = append(
			details,
			&ApplicationDetail{Application: application, Job: byID[application.JobID]},
		)
	}

	return details
}

func JobIDs(applications []*applicationModel.Application) []string {
	jobIDs := make([]string, 0, len(applications))
	for _, application := range applications {
		jobIDs = append(jobIDs, application.JobID)
	}

	return jobIDs
}
