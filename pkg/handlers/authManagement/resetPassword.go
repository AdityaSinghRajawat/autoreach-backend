package authManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var request authTypes.ResetPasswordRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	if customErr := h.authService.ResetPassword(c.Request.Context(), request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, nil, http.StatusNoContent)
}
