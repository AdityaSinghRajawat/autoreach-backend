package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader(config.GetOriginHeader())
		if origin == "" || !slices.Contains(allowedOrigins, origin) {
			c.Next()

			return
		}

		c.Header(config.GetCORSAllowOriginHeader(), origin)
		c.Header(config.GetCORSAllowCredentialsHeader(), config.GetTrueValue())
		c.Header(config.GetVaryHeader(), config.GetOriginHeader())

		if c.Request.Method != http.MethodOptions {
			c.Next()

			return
		}

		c.Header(config.GetCORSAllowMethodsHeader(), config.GetCORSAllowedMethods())
		c.Header(config.GetCORSAllowHeadersHeader(), config.GetCORSAllowedHeaders())
		c.Header(config.GetCORSMaxAgeHeader(), config.GetCORSMaxAgeSeconds())
		c.AbortWithStatus(http.StatusNoContent)
	}
}
