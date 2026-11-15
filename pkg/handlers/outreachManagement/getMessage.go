package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) GetMessage(c *gin.Context) {
	message, customErr := h.outreachService.GetMessageDetail(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetMessageIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, message, http.StatusOK)
}
