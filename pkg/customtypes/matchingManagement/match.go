package matchingManagement

import (
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
)

type ListMatchedJobsParams struct {
	sharedTypes.PageParams

	MinScore *float64 `form:"minScore" validate:"omitnil,gte=0,lte=1"`
}

type MatchedJob struct {
	JobID                   string    `json:"jobId"`
	CompanyName             string    `json:"companyName"`
	JobTitle                string    `json:"jobTitle"`
	Location                *string   `json:"location"`
	MatchScore              float64   `json:"matchScore"`
	ATSType                 string    `json:"atsType"`
	DiscoveredContactsCount int       `json:"discoveredContactsCount"`
	ApplicationStatus       *string   `json:"applicationStatus"`
	CreatedAt               time.Time `json:"createdAt"`
}

type ResumeEmbedding struct {
	ResumeID    string
	CandidateID string
	Vector      string
	IsActive    bool
}

type JobSimilarity struct {
	JobID      string
	Similarity float64
}

type ResumeSimilarity struct {
	ResumeID            string
	CandidateID         string
	MinMatchScore       float64
	Similarity          float64
	AutoOutreachEnabled bool
}

type MatchResult struct {
	Considered int
	Stored     int
}

func (s JobSimilarity) ToJobMatchModel(candidateID, resumeID string) *matchingModel.JobMatch {
	return newJobMatch(candidateID, s.JobID, resumeID, s.Similarity)
}

func (s ResumeSimilarity) ToJobMatchModel(jobID string) *matchingModel.JobMatch {
	return newJobMatch(s.CandidateID, jobID, s.ResumeID, s.Similarity)
}

func newJobMatch(candidateID, jobID, resumeID string, similarity float64) *matchingModel.JobMatch {
	match := matchingModel.NewJobMatch()
	match.CandidateID = candidateID
	match.JobID = jobID
	match.ResumeID = resumeID
	match.RawSimilarity = max(-1, min(1, similarity))
	match.MatchScore = config.CalibrateMatchScore(similarity)

	return match
}

func StorageThreshold(candidateMinMatchScore float64) float64 {
	return min(candidateMinMatchScore, config.GetMatchMinStoredScore())
}
