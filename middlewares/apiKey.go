package middlewares

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type APIKeyVerifier interface {
	VerifyAPIKey(ctx context.Context, apiKey, requiredScope string) *config.CustomError
}

func AuthenticateAPIClient(verifier APIKeyVerifier, requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(config.GetAPIKeyHeader())
		if apiKey == "" {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("authenticate api client: %w", config.ErrAPIKeyMissing),
					config.GetErrCodeUnauthorized(),
				),
			)

			return
		}

		if customErr := verifier.VerifyAPIKey(
			c.Request.Context(),
			apiKey,
			requiredScope,
		); customErr != nil {
			utils.HandleCustomError(c, customErr)

			return
		}

		c.Next()
	}
}
