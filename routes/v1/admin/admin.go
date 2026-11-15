package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadAdminRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	authHandler := deps.Handlers.Auth
	jobHandler := deps.Handlers.Job
	jobSourcePath := "/job-sources/:" + config.GetJobSourceIDParam()

	adminGroup := group.Group(
		"/admin",
		middlewares.AuthenticateUser(),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyUser),
	)
	adminGroup.POST(
		"/api-clients",
		middlewares.Authorize(config.PermissionAPIClientCreate),
		authHandler.CreateAPIClient,
	)
	adminGroup.GET(
		"/api-clients",
		middlewares.Authorize(config.PermissionAPIClientRead),
		authHandler.ListAPIClients,
	)
	adminGroup.DELETE(
		"/api-clients/:"+config.GetClientIDParam(),
		middlewares.Authorize(config.PermissionAPIClientRevoke),
		authHandler.RevokeAPIClient,
	)
	adminGroup.GET(
		"/job-sources",
		middlewares.Authorize(config.PermissionJobSourceRead),
		jobHandler.ListJobSources,
	)
	adminGroup.POST(
		"/job-sources",
		middlewares.Authorize(config.PermissionJobSourceCreate),
		jobHandler.CreateJobSource,
	)
	adminGroup.PATCH(
		jobSourcePath,
		middlewares.Authorize(config.PermissionJobSourceUpdate),
		jobHandler.UpdateJobSource,
	)
	adminGroup.DELETE(
		jobSourcePath,
		middlewares.Authorize(config.PermissionJobSourceDelete),
		jobHandler.DeleteJobSource,
	)
	adminGroup.POST(
		jobSourcePath+"/poll",
		middlewares.Authorize(config.PermissionJobSourceUpdate),
		jobHandler.PollJobSource,
	)
}
