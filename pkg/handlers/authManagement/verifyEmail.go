package authManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var request authTypes.VerifyEmailRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	if customErr := h.authService.VerifyEmail(c.Request.Context(), request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, nil, http.StatusNoContent)
}
