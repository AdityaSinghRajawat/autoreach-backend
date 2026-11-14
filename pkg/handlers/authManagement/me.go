package authManagement

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AuthHandler) Me(c *gin.Context) {
	principal, found := utils.GetPrincipal(c)
	if !found {
		utils.HandleCustomError(
			c,
			config.NewCustomError(
				fmt.Errorf("me: %w", config.ErrPrincipalMissing),
				config.GetErrCodeUnauthorized(),
			),
		)

		return
	}

	user, customErr := h.authService.GetCurrentUser(c.Request.Context(), principal)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, user, http.StatusOK)
}
