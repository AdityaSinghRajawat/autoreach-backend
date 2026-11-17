package cronJobs

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (cm *CronManager) dataRetentionJob(ctx context.Context) {
	interval := config.GetDataRetentionCheckInterval()

	cm.cronInstance.Schedule(cron.Every(interval), cron.FuncJob(func() {
		cm.runOncePerInterval(ctx, config.GetDataRetentionLock(), interval, cm.RunDataRetention)
	}))
}

func (cm *CronManager) RunDataRetention(ctx context.Context) {
	var anonymized, tokens, events int

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "anonymize stale contacts",
			Run: func(ctx context.Context) error {
				count, customErr := cm.outreachService.AnonymizeStaleContacts(
					ctx,
					config.GetContactRetentionPeriod(),
				)
				anonymized = count

				if customErr != nil {
					return customErr
				}

				return nil
			},
		},
		utils.ConcurrentTask{
			Name: "purge expired refresh tokens",
			Run: func(ctx context.Context) error {
				count, customErr := cm.authService.PurgeExpiredRefreshTokens(
					ctx,
					config.GetRefreshTokenRetention(),
				)
				tokens = count

				if customErr != nil {
					return customErr
				}

				return nil
			},
		},
		utils.ConcurrentTask{
			Name: "purge outreach events",
			Run: func(ctx context.Context) error {
				count, customErr := cm.outreachService.PurgeOutreachEvents(
					ctx,
					config.GetOutreachEventRetention(),
				)
				events = count

				if customErr != nil {
					return customErr
				}

				return nil
			},
		},
	)
	if err != nil {
		utils.LogError(ctx, err, "cron_job", config.GetDataRetentionLock())
	}

	utils.LogInfo(ctx, "data retention run", "contacts_anonymized", anonymized,
		"refresh_tokens_deleted", tokens, "outreach_events_deleted", events)
}
