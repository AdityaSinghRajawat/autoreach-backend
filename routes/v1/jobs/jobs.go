package jobs

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadJobRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	handler := deps.Handlers.Job

	jobGroup := group.Group("/jobs")
	jobGroup.POST(
		"/ingest",
		middlewares.AuthenticateAPIClient(deps.Services.Auth),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyAPIClient),
		middlewares.Authorize(config.PermissionJobIngest),
		handler.IngestJobs,
	)
	jobGroup.GET(
		"/:"+config.GetJobIDParam(),
		middlewares.AuthenticateUser(),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyUser),
		middlewares.Authorize(config.PermissionJobRead),
		handler.GetJob,
	)
}
