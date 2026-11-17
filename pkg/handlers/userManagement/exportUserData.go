package userManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *UserHandler) ExportUserData(c *gin.Context) {
	principal := utils.AuthenticatedPrincipal(c)

	export, customErr := h.userService.ExportUserData(c.Request.Context(), principal)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, export, http.StatusOK)
}
