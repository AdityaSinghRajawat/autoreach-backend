package dashboardManagement

import (
	"math"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// CandidateCounts is every count the dashboard reports for one candidate.
type CandidateCounts struct {
	TotalJobsScraped            int
	JobsMatched                 int
	ApplicationsSubmitted       int
	Interviews                  int
	ContactsDiscovered          int
	ColdEmailsSent              int
	LinkedinRequestsSent        int
	LinkedinConnectionsAccepted int
	RecruiterResponses          int
}

type MetricsSummary struct {
	TotalJobsScraped            int     `json:"totalJobsScraped"`
	JobsMatched                 int     `json:"jobsMatched"`
	ApplicationsSubmitted       int     `json:"applicationsSubmitted"`
	ContactsDiscovered          int     `json:"contactsDiscovered"`
	ColdEmailsSent              int     `json:"coldEmailsSent"`
	LinkedinRequestsSent        int     `json:"linkedinRequestsSent"`
	LinkedinConnectionsAccepted int     `json:"linkedinConnectionsAccepted"`
	RecruiterResponses          int     `json:"recruiterResponses"`
	ResponseRatePercentage      float64 `json:"responseRatePercentage"`
}

type CandidateMetricsResponse struct {
	CandidateID string         `json:"candidateId"`
	Metrics     MetricsSummary `json:"metrics"`
}

type FunnelStage struct {
	Stage string `json:"stage"`
	Count int    `json:"count"`
}

type FunnelResponse struct {
	CandidateID string        `json:"candidateId"`
	Stages      []FunnelStage `json:"stages"`
}

// ToMetricsResponse reports the response rate as recruiterResponses / coldEmailsSent x 100, rounded
// to two decimals (18 replies to 112 emails is 16.07).
func (c CandidateCounts) ToMetricsResponse(candidateID string) CandidateMetricsResponse {
	responseRate := 0.0
	if c.ColdEmailsSent > 0 {
		responseRate = math.Round(
			float64(c.RecruiterResponses)/float64(c.ColdEmailsSent)*config.GetPercentageScale()*100,
		) / 100
	}

	return CandidateMetricsResponse{
		CandidateID: candidateID,
		Metrics: MetricsSummary{
			TotalJobsScraped:            c.TotalJobsScraped,
			JobsMatched:                 c.JobsMatched,
			ApplicationsSubmitted:       c.ApplicationsSubmitted,
			ContactsDiscovered:          c.ContactsDiscovered,
			ColdEmailsSent:              c.ColdEmailsSent,
			LinkedinRequestsSent:        c.LinkedinRequestsSent,
			LinkedinConnectionsAccepted: c.LinkedinConnectionsAccepted,
			RecruiterResponses:          c.RecruiterResponses,
			ResponseRatePercentage:      responseRate,
		},
	}
}

func (c CandidateCounts) ToFunnelResponse(candidateID string) FunnelResponse {
	return FunnelResponse{
		CandidateID: candidateID,
		Stages: []FunnelStage{
			{Stage: "JOBS_SCRAPED", Count: c.TotalJobsScraped},
			{Stage: "JOBS_MATCHED", Count: c.JobsMatched},
			{Stage: "APPLICATIONS_SUBMITTED", Count: c.ApplicationsSubmitted},
			{Stage: "COLD_EMAILS_SENT", Count: c.ColdEmailsSent},
			{Stage: "LINKEDIN_REQUESTS_SENT", Count: c.LinkedinRequestsSent},
			{Stage: "RECRUITER_RESPONSES", Count: c.RecruiterResponses},
			{Stage: "INTERVIEWS", Count: c.Interviews},
		},
	}
}
