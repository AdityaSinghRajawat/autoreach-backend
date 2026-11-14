package candidateManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *CandidateHandler) GetCandidate(c *gin.Context) {
	candidate, customErr := h.candidateService.GetCandidate(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, candidate, http.StatusOK)
}
