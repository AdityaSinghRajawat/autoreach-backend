package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) CreateSuppression(c *gin.Context) {
	var request outreachTypes.CreateSuppressionRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	entry, customErr := h.outreachService.CreateSuppression(
		c.Request.Context(),
		request.ToSuppressionEntryModel(),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, entry, http.StatusCreated)
}
