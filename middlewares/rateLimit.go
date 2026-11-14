package middlewares

import (
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func RateLimit(
	limiter utils.Limiter,
	scope string,
	limit int,
	window time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		decision, err := limiter.Allow(
			ctx,
			scope+config.GetKeyDelimiter()+c.ClientIP(),
			limit,
			window,
		)
		if err != nil {
			utils.LogWarn(ctx, "rate limiter unavailable, allowing request", "error", err.Error())
			c.Next()

			return
		}

		if !decision.Allowed {
			retryAfterSeconds := int(math.Ceil(decision.RetryAfter.Seconds()))
			c.Header(config.GetRetryAfterHeader(), strconv.Itoa(retryAfterSeconds))
			utils.HandleCustomError(
				c,
				config.NewCustomError(config.ErrRateLimitExceeded, config.GetErrCodeRateLimited()),
			)

			return
		}

		c.Next()
	}
}
