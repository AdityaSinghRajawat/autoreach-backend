package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) DeleteTemplate(c *gin.Context) {
	if customErr := h.outreachService.DeleteTemplate(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetTemplateIDParam()),
	); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, nil, http.StatusNoContent)
}
