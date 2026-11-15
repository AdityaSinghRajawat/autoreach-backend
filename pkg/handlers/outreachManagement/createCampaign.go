package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) CreateCampaign(c *gin.Context) {
	var request outreachTypes.CreateCampaignRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	result, customErr := h.outreachService.CreateCampaign(
		c.Request.Context(),
		request.ToCampaignModel(c.Param(config.GetCandidateIDParam())),
		request.Steps,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusCreated)
}
