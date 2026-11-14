package enrichmentManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *EnrichmentHandler) ListJobContacts(c *gin.Context) {
	contacts, customErr := h.enrichmentService.ListJobContacts(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetJobIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, contacts, http.StatusOK)
}
