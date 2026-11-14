package enrichmentManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *EnrichmentHandler) RequestJobEnrichment(c *gin.Context) {
	response, customErr := h.enrichmentService.RequestJobEnrichment(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetJobIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, response, http.StatusAccepted)
}
