package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ListChannelAccounts(c *gin.Context) {
	accounts, customErr := h.outreachService.ListChannelAccounts(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, accounts, http.StatusOK)
}
