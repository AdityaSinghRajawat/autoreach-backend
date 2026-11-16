package adminManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AdminHandler) ListQueues(c *gin.Context) {
	result, customErr := h.adminService.ListQueues(c.Request.Context())
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
