package userManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	principal := utils.AuthenticatedPrincipal(c)

	user, customErr := h.userService.GetCurrentUser(c.Request.Context(), principal)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, user, http.StatusOK)
}
