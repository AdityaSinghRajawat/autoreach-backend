package enrichmentManagement

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *EnrichmentHandler) SignalHireWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.HandleCustomError(c, config.NewCustomError(
			fmt.Errorf("read signalhire webhook body: %w", err),
			config.GetErrCodeBadRequest(),
		))

		return
	}

	if customErr := h.enrichmentService.CompleteAsyncEnrichment(
		c.Request.Context(),
		config.ProviderSignalHire,
		c.GetHeader(config.GetSignalHireRequestIDHeader()),
		payload,
	); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, enrichmentTypes.WebhookAckResponse{Received: true}, http.StatusOK)
}
