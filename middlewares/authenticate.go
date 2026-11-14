package middlewares

import (
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, found := strings.CutPrefix(
			c.GetHeader(config.GetAuthorizationHeader()),
			config.GetBearerPrefix(),
		)
		if !found {
			utils.HandleCustomError(
				c,
				unauthorized("authenticate user", config.ErrBearerTokenMissing),
			)

			return
		}

		principal, err := utils.ParseAccessToken(token)
		if err != nil {
			utils.HandleCustomError(c, unauthorized("authenticate user", err))

			return
		}

		utils.SetPrincipal(c, principal)
		c.Next()
	}
}

func AuthenticateAPIClient(verifier sharedTypes.APIKeyVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(config.GetAPIKeyHeader())
		if apiKey == "" {
			utils.HandleCustomError(
				c,
				unauthorized("authenticate api client", config.ErrAPIKeyMissing),
			)

			return
		}

		principal, customErr := verifier.VerifyAPIKey(c.Request.Context(), apiKey)
		if customErr != nil {
			utils.HandleCustomError(c, customErr)

			return
		}

		utils.SetPrincipal(c, *principal)
		c.Next()
	}
}

func InternalSharedSecret(secret string) gin.HandlerFunc {
	return requireSecret(secret, config.ErrInternalSecretMismatch, func(c *gin.Context) string {
		return c.GetHeader(config.GetInternalSecretHeader())
	})
}

func WebhookToken(secret string) gin.HandlerFunc {
	return requireSecret(secret, config.ErrWebhookTokenMismatch, func(c *gin.Context) string {
		return c.Query(config.GetWebhookTokenParam())
	})
}

func requireSecret(
	secret string,
	mismatch error,
	provided func(c *gin.Context) string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" ||
			subtle.ConstantTimeCompare([]byte(provided(c)), []byte(secret)) != 1 {
			utils.HandleCustomError(c, unauthorized("authenticate caller", mismatch))

			return
		}

		c.Next()
	}
}

func unauthorized(operation string, err error) *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("%s: %w", operation, err),
		config.GetErrCodeUnauthorized(),
	)
}
