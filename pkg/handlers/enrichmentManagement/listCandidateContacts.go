package enrichmentManagement

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *EnrichmentHandler) ListCandidateContacts(c *gin.Context) {
	var params enrichmentTypes.ListContactsParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	contacts, meta, customErr := h.enrichmentService.ListCandidateContacts(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		params,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendPaginatedResponse(c, contacts, *meta)
}
