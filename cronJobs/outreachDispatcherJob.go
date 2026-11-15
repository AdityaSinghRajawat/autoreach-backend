package cronJobs

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (cm *CronManager) outreachDispatcherJob(ctx context.Context) {
	interval := config.GetOutreachDispatchInterval()

	cm.cronInstance.Schedule(cron.Every(interval), cron.FuncJob(func() {
		cm.runLocked(ctx, config.GetOutreachDispatchLock(), interval, cm.RunOutreachDispatcher)
	}))
}

func (cm *CronManager) RunOutreachDispatcher(ctx context.Context) {
	dispatched, customErr := cm.outreachService.DispatchDueMessages(ctx)
	if customErr != nil {
		utils.LogError(ctx, customErr)

		return
	}

	if dispatched > 0 {
		utils.LogInfo(ctx, "outreach dispatcher run", "dispatched", dispatched)
	}
}
