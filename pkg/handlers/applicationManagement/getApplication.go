package applicationManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	result, customErr := h.applicationService.GetApplication(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetApplicationIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
