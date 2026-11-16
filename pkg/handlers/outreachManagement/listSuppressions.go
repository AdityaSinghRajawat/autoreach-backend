package outreachManagement

import (
	"github.com/gin-gonic/gin"

	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ListSuppressions(c *gin.Context) {
	var params outreachTypes.ListSuppressionsParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	entries, meta, customErr := h.outreachService.ListSuppressions(c.Request.Context(), params)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendPaginatedResponse(c, entries, *meta)
}
