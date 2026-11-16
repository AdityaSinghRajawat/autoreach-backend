package dashboardManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *DashboardHandler) GetMetrics(c *gin.Context) {
	result, customErr := h.dashboardService.GetMetrics(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
