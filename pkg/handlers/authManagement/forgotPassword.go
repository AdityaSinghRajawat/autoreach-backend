package authManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var request authTypes.ForgotPasswordRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	if customErr := h.authService.ForgotPassword(c.Request.Context(), request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(
		c,
		authTypes.PasswordResetPendingResponse{ResetEmailSent: true},
		http.StatusAccepted,
	)
}
