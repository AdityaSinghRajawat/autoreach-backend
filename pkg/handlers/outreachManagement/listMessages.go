package outreachManagement

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ListMessages(c *gin.Context) {
	var params outreachTypes.ListMessagesParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	messages, meta, customErr := h.outreachService.ListMessages(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		params,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendPaginatedResponse(c, messages, *meta)
}
