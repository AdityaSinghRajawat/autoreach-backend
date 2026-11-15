package candidateManagement

import (
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type UploadResumeRequest struct {
	FileName string `json:"fileName" validate:"required,max=1024"`
	Content  []byte `json:"file"     validate:"required,min=1,validate_resume_file"`
}

type UpdateResumeRequest struct {
	IsActive *bool `json:"isActive" validate:"required,eq=true"`
}

type UploadResumeResponse struct {
	Resume   *candidateModel.Resume `json:"resume"`
	TaskName string                 `json:"taskName"`
	Status   string                 `json:"status"`
}

type ResumeProfile struct {
	Skills            []string          `json:"skills"`
	JobTitles         []string          `json:"jobTitles"`
	YearsOfExperience *float64          `json:"yearsOfExperience"`
	Education         []ResumeEducation `json:"education"`
}

type ResumeEducation struct {
	Degree         string `json:"degree,omitempty"`
	FieldOfStudy   string `json:"fieldOfStudy,omitempty"`
	Institution    string `json:"institution,omitempty"`
	GraduationYear *int   `json:"graduationYear,omitempty"`
}

type ResumeMetadata struct {
	ResumeProfile
	WordCount int    `json:"wordCount"`
	ParsedBy  string `json:"parsedBy"`
}

type ProcessResumeResult struct {
	ResumeID string
	Status   modelHelpers.ResumeStatus
}

func (r *UploadResumeRequest) SetFile(fileName string, content []byte) {
	r.FileName = fileName
	r.Content = content
}

func (r *UploadResumeRequest) ToResumeModel(candidateID string) *candidateModel.Resume {
	resume := candidateModel.NewResume()
	resume.CandidateID = candidateID
	resume.FileName = modelHelpers.SanitizeFileName(r.FileName)
	resume.ContentType = modelHelpers.DetectResumeContentType(r.Content)
	resume.SizeBytes = int64(len(r.Content))
	resume.ContentSHA256 = utils.Fingerprint(string(r.Content))
	resume.StorageKey = strings.Join([]string{
		config.GetResumeStoragePrefix(),
		candidateID,
		resume.ID + modelHelpers.ResumeExtension(resume.ContentType),
	}, "/")

	return resume
}

func NewUploadResumeResponse(
	resume *candidateModel.Resume,
	taskName, status string,
) *UploadResumeResponse {
	return &UploadResumeResponse{Resume: resume, TaskName: taskName, Status: status}
}
