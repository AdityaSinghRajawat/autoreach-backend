package adminManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AdminHandler) RetryDeadTask(c *gin.Context) {
	customErr := h.adminService.RetryDeadTask(
		c.Request.Context(),
		c.Param(config.GetQueueParam()),
		c.Param(config.GetTaskIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, nil, http.StatusNoContent)
}
