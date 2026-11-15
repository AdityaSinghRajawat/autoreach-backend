package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) CreateTemplate(c *gin.Context) {
	var request outreachTypes.CreateTemplateRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	result, customErr := h.outreachService.CreateTemplate(
		c.Request.Context(),
		request.ToTemplateModel(c.Param(config.GetCandidateIDParam())),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, result, http.StatusCreated)
}
