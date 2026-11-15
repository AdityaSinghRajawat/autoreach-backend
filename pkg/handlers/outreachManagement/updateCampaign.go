package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) UpdateCampaign(c *gin.Context) {
	var request outreachTypes.UpdateCampaignRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	result, customErr := h.outreachService.UpdateCampaign(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetCampaignIDParam()),
		request,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
