package candidates

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadCandidateRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	candidateHandler := deps.Handlers.Candidate
	matchingHandler := deps.Handlers.Matching
	enrichmentHandler := deps.Handlers.Enrichment
	jobContactsPath := "/jobs/:" + config.GetJobIDParam() + "/contacts"
	resumePath := "/resumes/:" + config.GetResumeIDParam()

	candidateGroup := group.Group(
		"/candidates/:"+config.GetCandidateIDParam(),
		middlewares.AuthenticateUser(),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyUser),
	)
	candidateGroup.GET(
		"",
		middlewares.Authorize(config.PermissionCandidateRead),
		candidateHandler.GetCandidate,
	)
	candidateGroup.PATCH(
		"",
		middlewares.Authorize(config.PermissionCandidateUpdate),
		candidateHandler.UpdateCandidate,
	)

	candidateGroup.POST(
		"/resumes",
		middlewares.Authorize(config.PermissionResumeCreate),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyResumeUpload),
		candidateHandler.UploadResume,
	)
	candidateGroup.GET(
		"/resumes",
		middlewares.Authorize(config.PermissionResumeRead),
		candidateHandler.ListResumes,
	)
	candidateGroup.GET(
		resumePath,
		middlewares.Authorize(config.PermissionResumeRead),
		candidateHandler.GetResume,
	)
	candidateGroup.PATCH(
		resumePath,
		middlewares.Authorize(config.PermissionResumeUpdate),
		candidateHandler.UpdateResume,
	)
	candidateGroup.DELETE(
		resumePath,
		middlewares.Authorize(config.PermissionResumeDelete),
		candidateHandler.DeleteResume,
	)

	candidateGroup.GET(
		"/jobs",
		middlewares.Authorize(config.PermissionMatchRead),
		matchingHandler.ListMatchedJobs,
	)

	candidateGroup.GET(
		"/contacts",
		middlewares.Authorize(config.PermissionContactRead),
		enrichmentHandler.ListCandidateContacts,
	)
	candidateGroup.GET(
		jobContactsPath,
		middlewares.Authorize(config.PermissionContactRead),
		enrichmentHandler.ListJobContacts,
	)
	candidateGroup.POST(
		jobContactsPath+"/enrich",
		middlewares.Authorize(config.PermissionContactEnrich),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyContactEnrich),
		enrichmentHandler.RequestJobEnrichment,
	)
}
