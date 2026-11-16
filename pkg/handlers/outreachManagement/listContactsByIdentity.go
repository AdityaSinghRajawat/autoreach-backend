package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ListContactsByIdentity(c *gin.Context) {
	var params outreachTypes.ContactLookupParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	contacts, customErr := h.outreachService.ListContactsByIdentity(c.Request.Context(), params)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(c, contacts, http.StatusOK)
}
