package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ListContactSignals(c *gin.Context) {
	signals, customErr := h.outreachService.ListContactSignals(
		c.Request.Context(),
		c.Param(config.GetContactIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, signals, http.StatusOK)
}
