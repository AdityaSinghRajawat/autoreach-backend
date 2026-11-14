package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		customErr := toCustomError(c.Errors.Last().Err)
		logHandledError(c, customErr)

		if c.Writer.Written() {
			return
		}

		utils.SendErrorResponse(c, customErr)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		customErr := config.NewCustomError(
			fmt.Errorf("panic recovered: %v", recovered),
			config.GetErrCodeInternalServerError(),
		)
		logHandledError(c, customErr)
		utils.SendErrorResponse(c, customErr)
	})
}

func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.HandleCustomError(
			c,
			config.NewCustomError(config.ErrRouteNotFound, config.GetErrCodeNotFound()),
		)
	}
}

func MethodNotAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.HandleCustomError(
			c,
			config.NewCustomError(config.ErrMethodNotAllowed, config.GetErrCodeMethodNotAllowed()),
		)
	}
}

func toCustomError(err error) *config.CustomError {
	var customErr *config.CustomError
	if errors.As(err, &customErr) {
		return customErr
	}

	return config.NewCustomError(err, config.GetErrCodeInternalServerError())
}

func logHandledError(c *gin.Context, customErr *config.CustomError) {
	ctx := c.Request.Context()

	status := utils.GetHTTPStatusForErrCode(customErr.ErrCode)
	if status >= http.StatusInternalServerError {
		utils.LogError(ctx, customErr, "err_code", string(customErr.ErrCode), "path", c.FullPath())

		return
	}

	utils.LogWarn(
		ctx,
		customErr.Error(),
		"err_code",
		string(customErr.ErrCode),
		"path",
		c.FullPath(),
	)
}
