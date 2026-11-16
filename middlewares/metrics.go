package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Metrics records every request under its route template, so IDs in paths do not create new series.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = config.GetUnmatchedRoute()
		}

		utils.RecordHTTPRequest(c.Request.Method, route, responseStatus(c), time.Since(startedAt))
	}
}
