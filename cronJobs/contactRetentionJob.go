package cronJobs

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (cm *CronManager) contactRetentionJob(ctx context.Context) {
	if config.GetContactRetentionPeriod() <= 0 {
		return
	}

	interval := config.GetContactRetentionCheckInterval()

	cm.cronInstance.Schedule(cron.Every(interval), cron.FuncJob(func() {
		cm.runOncePerInterval(
			ctx,
			config.GetContactRetentionLock(),
			interval,
			cm.RunContactRetention,
		)
	}))
}

func (cm *CronManager) RunContactRetention(ctx context.Context) {
	anonymized, customErr := cm.outreachService.AnonymizeStaleContacts(
		ctx,
		config.GetContactRetentionPeriod(),
	)
	if customErr != nil {
		utils.LogError(ctx, customErr)

		return
	}

	utils.LogInfo(ctx, "stale contacts anonymized", "contacts", anonymized)
}
