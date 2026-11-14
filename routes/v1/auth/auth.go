package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadAuthRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	handler := deps.Handlers.Auth
	attemptLimit := middlewares.RateLimit(
		deps.RateLimiter,
		config.GetRateLimitScopeAuth(),
		config.GetAuthRateLimitRequests(),
		config.GetAuthRateLimitWindow(),
	)

	authGroup := group.Group("/auth")
	authGroup.POST("/register", attemptLimit, handler.Register)
	authGroup.POST("/login", attemptLimit, handler.Login)
	authGroup.POST("/refresh", attemptLimit, handler.Refresh)
	authGroup.POST("/logout", handler.Logout)
	authGroup.GET("/me", middlewares.AuthenticateUser(), handler.Me)
}
