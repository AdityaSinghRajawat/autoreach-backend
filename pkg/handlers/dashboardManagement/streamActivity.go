package dashboardManagement

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Each event id is its cursor, so a reconnecting client resumes from Last-Event-ID.
func (h *DashboardHandler) StreamActivity(c *gin.Context) {
	controller := http.NewResponseController(c.Writer)
	started := false

	ctx, stop := utils.UntilShutdown(c.Request.Context())
	defer stop()

	customErr := h.dashboardService.StreamActivity(
		ctx,
		c.Param(config.GetCandidateIDParam()),
		c.GetHeader(config.GetLastEventIDHeader()),
		func(items []dashboardTypes.ActivityItem) error {
			if !started {
				started = true

				c.Header(config.GetContentTypeHeader(), config.GetSSEContentType())
				c.Header(config.GetCacheControlHeader(), config.GetNoCacheValue())
				c.Header(config.GetProxyBufferingHeader(), config.GetProxyBufferingOff())
				c.Status(http.StatusOK)
			}

			_ = controller.SetWriteDeadline(time.Now().Add(config.GetActivityStreamHeartbeat() * 2))

			if len(items) == 0 {
				if _, err := fmt.Fprint(c.Writer, ": keep-alive\n\n"); err != nil {
					return err
				}
			}

			for _, item := range items {
				data, err := json.Marshal(item)
				if err != nil {
					return err
				}

				if _, err := fmt.Fprintf(
					c.Writer,
					"id: %s\nevent: %s\ndata: %s\n\n",
					item.Cursor().Encode(),
					config.GetSSEActivityEvent(),
					data,
				); err != nil {
					return err
				}
			}

			return controller.Flush()
		},
	)
	if customErr != nil && !started {
		utils.HandleCustomError(c, customErr)
	}
}
