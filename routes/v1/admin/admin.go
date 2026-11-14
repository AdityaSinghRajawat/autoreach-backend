package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

func LoadAdminRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	authHandler := deps.Handlers.Auth

	adminGroup := group.Group(
		"/admin",
		middlewares.AuthenticateUser(),
		middlewares.RequireRole(modelHelpers.UserRoleAdmin),
	)
	adminGroup.POST("/api-clients", authHandler.CreateAPIClient)
	adminGroup.GET("/api-clients", authHandler.ListAPIClients)
	adminGroup.DELETE("/api-clients/:"+config.GetClientIDParam(), authHandler.RevokeAPIClient)
}
