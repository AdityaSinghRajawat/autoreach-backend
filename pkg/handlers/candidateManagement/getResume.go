package candidateManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *CandidateHandler) GetResume(c *gin.Context) {
	resume, customErr := h.candidateService.GetResume(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetResumeIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, resume, http.StatusOK)
}
