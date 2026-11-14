package candidateManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *CandidateHandler) DeleteResume(c *gin.Context) {
	if customErr := h.candidateService.DeleteResume(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetResumeIDParam()),
	); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, nil, http.StatusNoContent)
}
