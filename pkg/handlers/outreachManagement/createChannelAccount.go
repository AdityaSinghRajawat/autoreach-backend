package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) CreateChannelAccount(c *gin.Context) {
	var request outreachTypes.CreateChannelAccountRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	account, customErr := h.outreachService.CreateChannelAccount(
		c.Request.Context(),
		request.ToChannelAccountModel(c.Param(config.GetCandidateIDParam())),
		request.Credentials(),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, account, http.StatusCreated)
}
