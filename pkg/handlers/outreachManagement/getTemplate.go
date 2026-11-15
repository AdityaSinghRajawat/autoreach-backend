package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) GetTemplate(c *gin.Context) {
	result, customErr := h.outreachService.GetTemplate(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetTemplateIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
