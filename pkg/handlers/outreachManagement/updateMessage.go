package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) UpdateMessage(c *gin.Context) {
	var request outreachTypes.UpdateMessageRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	message, customErr := h.outreachService.UpdateMessage(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetMessageIDParam()),
		request,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, message, http.StatusOK)
}
