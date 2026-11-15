package jobManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *JobHandler) CreateJobSource(c *gin.Context) {
	var request jobTypes.CreateJobSourceRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	result, customErr := h.jobService.CreateJobSource(
		c.Request.Context(),
		request.ToJobSourceModel(),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusCreated)
}
