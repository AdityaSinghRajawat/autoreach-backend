package applicationManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *ApplicationHandler) GetApplicationProfile(c *gin.Context) {
	result, customErr := h.applicationService.GetApplicationProfile(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusOK)
}
