package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) TrackClick(c *gin.Context) {
	link, customErr := h.outreachService.RecordClick(
		c.Request.Context(),
		c.Param(config.GetTrackingTokenParam()),
		c.Query(config.GetTrackingURLParam()),
		c.Query(config.GetTrackingSignatureParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	c.Redirect(http.StatusFound, link)
}
