package jobManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *JobHandler) IngestJobs(c *gin.Context) {
	var request jobTypes.IngestJobsRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	response, customErr := h.jobService.IngestJobs(
		c.Request.Context(),
		request.ToJobPostingModels(),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, response, http.StatusAccepted)
}
