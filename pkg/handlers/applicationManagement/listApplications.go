package applicationManagement

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	var params applicationTypes.ListApplicationsParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	applications, meta, customErr := h.applicationService.ListApplications(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		params,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendPaginatedResponse(c, applications, *meta)
}
