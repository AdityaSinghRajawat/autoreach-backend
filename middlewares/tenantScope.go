package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func TenantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, found := utils.GetPrincipal(c)
		if !found {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("tenant scope: %w", config.ErrPrincipalMissing),
					config.GetErrCodeUnauthorized(),
				),
			)

			return
		}

		candidateID := c.Param(config.GetCandidateIDParam())
		if _, err := uuid.Parse(
			candidateID,
		); err != nil ||
			!principal.CanAccessCandidate(candidateID) {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("tenant scope: %w", config.ErrCandidateNotAccessible),
					config.GetErrCodeNotFound(),
				),
			)

			return
		}

		c.Next()
	}
}
