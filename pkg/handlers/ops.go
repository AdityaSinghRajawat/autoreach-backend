package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func Health(c *gin.Context) {
	utils.BuildAndSendResponse(c, sharedTypes.HealthResponse{Status: "ok"}, http.StatusOK)
}
