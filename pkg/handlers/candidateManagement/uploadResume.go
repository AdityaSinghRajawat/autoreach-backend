package candidateManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *CandidateHandler) UploadResume(c *gin.Context) {
	var request candidateTypes.UploadResumeRequest
	if customErr := utils.BindFileAndValidate(
		c,
		config.GetResumeFormField(),
		config.GetResumeMaxBytes(),
		&request,
	); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	response, customErr := h.candidateService.UploadResume(
		c.Request.Context(),
		request.ToResumeModel(c.Param(config.GetCandidateIDParam())),
		request.Content,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, response, http.StatusAccepted)
}
