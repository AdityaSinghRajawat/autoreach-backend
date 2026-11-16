package cronJobs

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (cm *CronManager) replySweeperJob(ctx context.Context) {
	interval := config.GetReplySweepInterval()

	cm.cronInstance.Schedule(cron.Every(interval), cron.FuncJob(func() {
		cm.runOncePerInterval(ctx, config.GetReplySweeperLock(), interval, cm.RunReplySweeper)
	}))
}

func (cm *CronManager) RunReplySweeper(ctx context.Context) {
	if customErr := cm.outreachService.ScheduleMailboxSweeps(ctx); customErr != nil {
		utils.LogError(ctx, customErr)
	}
}
