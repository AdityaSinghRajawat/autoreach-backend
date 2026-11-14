package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadAdminRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	authHandler := deps.Handlers.Auth

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
}
