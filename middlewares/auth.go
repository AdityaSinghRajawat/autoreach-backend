package middlewares

import (
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func InternalSharedSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader(config.GetInternalSecretHeader())
		if secret == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					config.ErrInternalSecretMismatch,
					config.GetErrCodeUnauthorized(),
				),
			)

			return
		}

		c.Next()
	}
}

func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(config.GetAuthorizationHeader())
		if !strings.HasPrefix(header, config.GetBearerPrefix()) {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("authenticate user: %w", config.ErrBearerTokenMissing),
					config.GetErrCodeUnauthorized(),
				),
			)

			return
		}

		principal, err := utils.ParseAccessToken(
			strings.TrimPrefix(header, config.GetBearerPrefix()),
		)
		if err != nil {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("authenticate user: %w", err),
					config.GetErrCodeUnauthorized(),
				),
			)

			return
		}

		utils.SetPrincipal(c, principal)
		c.Next()
	}
}
