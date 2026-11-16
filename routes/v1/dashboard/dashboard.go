package dashboard

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

// Dashboard routes serve the signed-in user's candidate; admins pass ?candidateId=.
func LoadDashboardRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	dashboardHandler := deps.Handlers.Dashboard

	dashboardGroup := group.Group(
		"/dashboard",
		middlewares.AuthenticateUser(),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyUser),
		middlewares.UsePrincipalCandidate(),
		middlewares.Authorize(config.PermissionDashboardRead),
	)
	dashboardGroup.GET("/funnel", dashboardHandler.GetFunnel)
	dashboardGroup.GET("/activity", dashboardHandler.ListActivity)
	dashboardGroup.GET("/activity/stream", dashboardHandler.StreamActivity)
}
