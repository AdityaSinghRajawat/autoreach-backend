package cronJobs

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (cm *CronManager) jobSourcePollerJob(ctx context.Context) {
	interval := config.GetJobSourcePollInterval()

	cm.cronInstance.Schedule(cron.Every(interval), cron.FuncJob(func() {
		cm.runLocked(ctx, config.GetJobSourcePollerLock(), interval, cm.RunJobSourcePoller)
	}))
}

func (cm *CronManager) RunJobSourcePoller(ctx context.Context) {
	scheduled, customErr := cm.jobService.ScheduleJobSourcePolls(ctx)
	if customErr != nil {
		utils.LogError(ctx, customErr)

		return
	}

	utils.LogInfo(ctx, "job source polls scheduled", "sources", scheduled)
}
