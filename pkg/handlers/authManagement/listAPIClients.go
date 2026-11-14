package authManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AuthHandler) ListAPIClients(c *gin.Context) {
	clients, customErr := h.authService.ListAPIClients(c.Request.Context())
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, clients, http.StatusOK)
}
