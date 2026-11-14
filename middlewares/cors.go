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

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Vary", config.GetOriginHeader())

		if c.Request.Method != http.MethodOptions {
			c.Next()

			return
		}

		c.Header("Access-Control-Allow-Methods", config.GetCORSAllowedMethods())
		c.Header("Access-Control-Allow-Headers", config.GetCORSAllowedHeaders())
		c.Header("Access-Control-Max-Age", config.GetCORSMaxAgeSeconds())
		c.AbortWithStatus(http.StatusNoContent)
	}
}
