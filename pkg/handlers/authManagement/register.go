package authManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AuthHandler) Register(c *gin.Context) {
	var request authTypes.RegisterRequest
	if customErr := utils.BindAndValidate(c, &request); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	user := request.ToUserModel()
	candidate := request.ToCandidateModel(user)

	session, customErr := h.authService.Register(
		c.Request.Context(),
		user,
		candidate,
		request.Password,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	if session == nil {
		utils.BuildAndSendResponse(
			c,
			authTypes.RegistrationPendingResponse{VerificationRequired: true},
			http.StatusAccepted,
		)

		return
	}

	utils.BuildAndSendResponse(c, session, http.StatusCreated)
}
