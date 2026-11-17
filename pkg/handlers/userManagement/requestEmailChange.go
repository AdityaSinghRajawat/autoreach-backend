package userManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *UserHandler) RequestEmailChange(c *gin.Context) {
	principal := utils.AuthenticatedPrincipal(c)

	var request userTypes.ChangeEmailRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	if customErr := h.userService.RequestEmailChange(
		c.Request.Context(),
		principal,
		request,
	); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendResponse(
		c,
		userTypes.EmailChangePendingResponse{ConfirmationSent: true},
		http.StatusAccepted,
	)
}
