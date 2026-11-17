package users

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadUserRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	handler := deps.Handlers.User
	attemptLimit := middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyAuth)

	currentUser := group.Group(
		"/users/me",
		middlewares.AuthenticateUser(),
		middlewares.RateLimit(deps.RateLimiter, config.RateLimitPolicyUser),
	)
	currentUser.GET("", handler.GetCurrentUser)
	currentUser.DELETE("", attemptLimit, handler.DeleteAccount)
	currentUser.GET("/export", handler.ExportUserData)
	currentUser.PUT("/password", attemptLimit, handler.ChangePassword)
	currentUser.POST("/email", attemptLimit, handler.RequestEmailChange)
	currentUser.POST("/email/confirm", attemptLimit, handler.ConfirmEmailChange)
}
