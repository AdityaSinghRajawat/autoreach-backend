package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func RequireRole(role modelHelpers.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, found := utils.GetPrincipal(c)
		if !found {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("require role: %w", config.ErrPrincipalMissing),
					config.GetErrCodeUnauthorized(),
				),
			)

			return
		}

		if principal.Role != role {
			utils.HandleCustomError(
				c,
				config.NewCustomError(
					fmt.Errorf("require role: %w", config.ErrInsufficientRole),
					config.GetErrCodeForbidden(),
				),
			)

			return
		}

		c.Next()
	}
}
