package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

var statusByErrCodeI = map[config.CustomErrorCode]int{
	config.GetErrCodeInternalServerError(): http.StatusInternalServerError,
	config.GetErrCodeBadRequest():          http.StatusBadRequest,
	config.GetErrCodeValidation():          http.StatusUnprocessableEntity,
	config.GetErrCodeUnauthorized():        http.StatusUnauthorized,
	config.GetErrCodeForbidden():           http.StatusForbidden,
	config.GetErrCodeNotFound():            http.StatusNotFound,
	config.GetErrCodeMethodNotAllowed():    http.StatusMethodNotAllowed,
	config.GetErrCodeConflict():            http.StatusConflict,
	config.GetErrCodePayloadTooLarge():     http.StatusRequestEntityTooLarge,
	config.GetErrCodeRateLimited():         http.StatusTooManyRequests,
	config.GetErrCodeServiceUnavailable():  http.StatusServiceUnavailable,
	config.GetErrCodeUpstreamRejected():    http.StatusBadGateway,
}

func GetHTTPStatusForErrCode(errCode config.CustomErrorCode) int {
	status, found := statusByErrCodeI[errCode]
	if !found {
		return http.StatusInternalServerError
	}

	return status
}

func IsErrCodeMapped(errCode config.CustomErrorCode) bool {
	_, found := statusByErrCodeI[errCode]

	return found
}

func HandleCustomError(c *gin.Context, customErr *config.CustomError) {
	_ = c.Error(customErr)
	c.Abort()
}

func SendErrorResponse(c *gin.Context, customErr *config.CustomError) {
	status := GetHTTPStatusForErrCode(customErr.ErrCode)
	c.AbortWithStatusJSON(status, sharedTypes.ServerResponse{
		Success:    false,
		StatusCode: status,
		Error: &sharedTypes.ErrorBody{
			Code:    string(customErr.ErrCode),
			Message: config.GetPublicMessageForErrCode(customErr.ErrCode),
			Details: customErr.Details,
		},
	})
}

func BuildAndSendResponse(c *gin.Context, data any, status int) {
	if status == http.StatusNoContent {
		c.Status(status)

		return
	}

	c.JSON(status, sharedTypes.ServerResponse{
		Success:    true,
		StatusCode: status,
		Data:       data,
	})
}

func BuildAndSendPaginatedResponse(c *gin.Context, data any, meta sharedTypes.PaginationMeta) {
	c.JSON(http.StatusOK, sharedTypes.ServerResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       data,
		Meta:       &meta,
	})
}
