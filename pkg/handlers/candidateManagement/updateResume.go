package candidateManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *CandidateHandler) UpdateResume(c *gin.Context) {
	var request candidateTypes.UpdateResumeRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	resume, customErr := h.candidateService.ActivateResume(
		c.Request.Context(),
		c.Param(config.GetCandidateIDParam()),
		c.Param(config.GetResumeIDParam()),
		request,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, resume, http.StatusOK)
}
