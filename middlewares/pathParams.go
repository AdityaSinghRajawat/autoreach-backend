package middlewares

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func ValidateIDParams() gin.HandlerFunc {
	uuidParams := config.GetUUIDPathParams()

	return func(c *gin.Context) {
		for _, param := range c.Params {
			if slices.Contains(uuidParams, param.Key) && uuid.Validate(param.Value) != nil {
				utils.HandleCustomError(c, config.NewCustomError(
					fmt.Errorf("path parameter %s: %w", param.Key, config.ErrResourceNotAccessible),
					config.GetErrCodeNotFound(),
				))

				return
			}
		}

		c.Next()
	}
}
