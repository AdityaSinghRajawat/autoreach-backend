package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		correlationID := c.GetHeader(config.GetCorrelationIDHeader())
		if _, err := uuid.Parse(correlationID); err != nil {
			correlationID = uuid.NewString()
		}

		requestID := uuid.NewString()

		ctx := utils.WithLogger(
			c.Request.Context(),
			logger.With(slog.String("request_id", requestID)),
		)
		ctx = utils.WithCorrelationID(ctx, correlationID)
		ctx = utils.WithRequestID(ctx, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(config.GetCorrelationIDHeader(), correlationID)

		c.Next()

		utils.LogInfo(
			ctx,
			"request completed",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
