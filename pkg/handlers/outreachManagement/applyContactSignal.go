package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ApplyContactSignal(c *gin.Context) {
	signal, customErr := h.outreachService.ApplyContactSignal(
		c.Request.Context(),
		c.Param(config.GetContactIDParam()),
		c.Param(config.GetContactSignalIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, signal, http.StatusOK)
}
