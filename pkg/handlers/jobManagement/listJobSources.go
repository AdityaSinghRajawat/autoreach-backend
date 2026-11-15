package jobManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *JobHandler) ListJobSources(c *gin.Context) {
	result, customErr := h.jobService.ListJobSources(c.Request.Context())
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
