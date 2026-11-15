package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) VerifyChannelAccount(c *gin.Context) {
	account, customErr := h.outreachService.VerifyChannelAccount(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetAccountIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, account, http.StatusOK)
}
