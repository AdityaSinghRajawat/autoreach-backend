package matchingManagement

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *MatchingHandler) ListMatchedJobs(c *gin.Context) {
	var params matchingTypes.ListMatchedJobsParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	jobs, meta, customErr := h.matchingService.ListMatchedJobs(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		params,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendPaginatedResponse(c, jobs, *meta)
}
