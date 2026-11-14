package jobManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *JobHandler) GetJob(c *gin.Context) {
	job, customErr := h.jobService.GetJob(c.Request.Context(), c.Param(config.GetJobIDParam()))
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, job, http.StatusOK)
}
