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
	outreachHandler := deps.Handlers.Outreach
	applicationHandler := deps.Handlers.Application
	jobContactsPath := "/jobs/:" + config.GetJobIDParam() + "/contacts"
	resumePath := "/resumes/:" + config.GetResumeIDParam()
	channelAccountPath := "/channel-accounts/:" + config.GetAccountIDParam()
	templatePath := "/outreach-templates/:" + config.GetTemplateIDParam()
	campaignPath := "/outreach-campaigns/:" + config.GetCampaignIDParam()
	messagePath := "/outreach-messages/:" + config.GetMessageIDParam()
	applicationPath := "/applications/:" + config.GetApplicationIDParam()

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

	candidateGroup.GET(
		"/channel-accounts",
		middlewares.Authorize(config.PermissionChannelAccountRead),
		outreachHandler.ListChannelAccounts,
	)
	candidateGroup.POST(
		"/channel-accounts",
		middlewares.Authorize(config.PermissionChannelAccountCreate),
		outreachHandler.CreateChannelAccount,
	)
	candidateGroup.DELETE(
		channelAccountPath,
		middlewares.Authorize(config.PermissionChannelAccountDelete),
		outreachHandler.DeleteChannelAccount,
	)
	candidateGroup.POST(
		channelAccountPath+"/verify",
		middlewares.Authorize(config.PermissionChannelAccountUpdate),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyMailboxVerify),
		outreachHandler.VerifyChannelAccount,
	)
	candidateGroup.GET(
		"/outreach-templates",
		middlewares.Authorize(config.PermissionOutreachTemplateRead),
		outreachHandler.ListTemplates,
	)
	candidateGroup.POST(
		"/outreach-templates",
		middlewares.Authorize(config.PermissionOutreachTemplateCreate),
		outreachHandler.CreateTemplate,
	)
	candidateGroup.GET(
		templatePath,
		middlewares.Authorize(config.PermissionOutreachTemplateRead),
		outreachHandler.GetTemplate,
	)
	candidateGroup.PATCH(
		templatePath,
		middlewares.Authorize(config.PermissionOutreachTemplateUpdate),
		outreachHandler.UpdateTemplate,
	)
	candidateGroup.DELETE(
		templatePath,
		middlewares.Authorize(config.PermissionOutreachTemplateDelete),
		outreachHandler.DeleteTemplate,
	)
	candidateGroup.GET(
		"/outreach-campaigns",
		middlewares.Authorize(config.PermissionOutreachCampaignRead),
		outreachHandler.ListCampaigns,
	)
	candidateGroup.POST(
		"/outreach-campaigns",
		middlewares.Authorize(config.PermissionOutreachCampaignCreate),
		outreachHandler.CreateCampaign,
	)
	candidateGroup.GET(
		campaignPath,
		middlewares.Authorize(config.PermissionOutreachCampaignRead),
		outreachHandler.GetCampaign,
	)
	candidateGroup.PATCH(
		campaignPath,
		middlewares.Authorize(config.PermissionOutreachCampaignUpdate),
		outreachHandler.UpdateCampaign,
	)
	candidateGroup.GET(
		"/outreach-messages",
		middlewares.Authorize(config.PermissionOutreachMessageRead),
		outreachHandler.ListMessages,
	)
	candidateGroup.POST(
		"/outreach-messages",
		middlewares.Authorize(config.PermissionOutreachMessageCreate),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyOutreachCreate),
		outreachHandler.CreateMessages,
	)
	candidateGroup.GET(
		messagePath,
		middlewares.Authorize(config.PermissionOutreachMessageRead),
		outreachHandler.GetMessage,
	)
	candidateGroup.PATCH(
		messagePath,
		middlewares.Authorize(config.PermissionOutreachMessageUpdate),
		outreachHandler.UpdateMessage,
	)
	candidateGroup.POST(
		messagePath+"/confirm",
		middlewares.Authorize(config.PermissionOutreachMessageUpdate),
		outreachHandler.ConfirmMessage,
	)
	candidateGroup.GET(
		"/application-profile",
		middlewares.Authorize(config.PermissionCandidateRead),
		applicationHandler.GetApplicationProfile,
	)
	candidateGroup.PUT(
		"/application-profile",
		middlewares.Authorize(config.PermissionCandidateUpdate),
		applicationHandler.UpdateApplicationProfile,
	)
	candidateGroup.GET(
		"/applications",
		middlewares.Authorize(config.PermissionApplicationRead),
		applicationHandler.ListApplications,
	)
	candidateGroup.POST(
		"/applications",
		middlewares.Authorize(config.PermissionApplicationCreate),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyApplicationCreate),
		applicationHandler.CreateApplication,
	)
	candidateGroup.GET(
		applicationPath,
		middlewares.Authorize(config.PermissionApplicationRead),
		applicationHandler.GetApplication,
	)
	candidateGroup.PATCH(
		applicationPath,
		middlewares.Authorize(config.PermissionApplicationUpdate),
		applicationHandler.UpdateApplication,
	)
}
