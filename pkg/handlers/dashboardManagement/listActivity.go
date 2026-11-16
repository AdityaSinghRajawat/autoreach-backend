package dashboardManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *DashboardHandler) ListActivity(c *gin.Context) {
	var params dashboardTypes.ListActivityParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	page, customErr := h.dashboardService.ListActivity(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		params,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, page, http.StatusOK)
}
