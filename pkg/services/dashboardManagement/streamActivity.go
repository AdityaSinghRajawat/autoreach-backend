package dashboardManagement

import (
	"context"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// StreamActivity calls emit with each batch of new events, and with no events as a heartbeat,
// until the client goes away, emit fails or ACTIVITY_STREAM_MAX_DURATION passes (clients reconnect
// with Last-Event-ID). Without a lastEventID the stream starts from now.
func (s *DashboardService) StreamActivity(
	ctx context.Context,
	candidateID, lastEventID string,
	emit func(items []dashboardTypes.ActivityItem) error,
) *config.CustomError {
	if candidateID == "" {
		return candidateNotFound()
	}

	cursor := dashboardTypes.ActivityCursor{OccurredAt: utils.GetCurrTimeStamp()}

	if lastEventID != "" {
		parsed, valid := dashboardTypes.ParseActivityCursor(lastEventID)
		if !valid {
			return invalidCursor()
		}

		cursor = parsed
	}

	if err := emit(nil); err != nil {
		return nil
	}

	poll := time.NewTicker(config.GetActivityStreamPollInterval())
	defer poll.Stop()

	heartbeat := time.NewTicker(config.GetActivityStreamHeartbeat())
	defer heartbeat.Stop()

	deadline := time.NewTimer(config.GetActivityStreamMaxDuration())
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-deadline.C:
			return nil
		case <-heartbeat.C:
			if err := emit(nil); err != nil {
				return nil
			}
		case <-poll.C:
			items, err := s.activityEventRepo.ListAfter(
				ctx,
				candidateID,
				cursor,
				config.GetActivityStreamBatchSize(),
			)
			if err != nil {
				utils.LogWarn(ctx, "activity stream poll failed", "error", err.Error())

				continue
			}

			if len(items) == 0 {
				continue
			}

			cursor = items[len(items)-1].Cursor()

			if err := emit(items); err != nil {
				return nil
			}
		}
	}
}
