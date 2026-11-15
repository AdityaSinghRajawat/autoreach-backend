package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) DeleteChannelAccount(c *gin.Context) {
	customErr := h.outreachService.DeleteChannelAccount(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetAccountIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, nil, http.StatusNoContent)
}
