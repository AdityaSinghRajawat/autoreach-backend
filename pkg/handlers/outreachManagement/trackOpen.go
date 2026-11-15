package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (h *OutreachHandler) TrackOpen(c *gin.Context) {
	h.outreachService.RecordOpen(c.Request.Context(), c.Param(config.GetTrackingTokenParam()))

	c.Header(config.GetCacheControlHeader(), config.GetNoStoreValue())
	c.Data(http.StatusOK, config.GetGIFContentType(), config.GetTrackingPixel())
}
