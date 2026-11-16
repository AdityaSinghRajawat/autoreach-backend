package cronJobs

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (cm *CronManager) stuckWorkReconcilerJob(ctx context.Context) {
	interval := config.GetReconcilerInterval()

	cm.cronInstance.Schedule(cron.Every(interval), cron.FuncJob(func() {
		cm.runOncePerInterval(
			ctx,
			config.GetStuckWorkReconcilerLock(),
			interval,
			cm.RunStuckWorkReconciler,
		)
	}))
}

func (cm *CronManager) RunStuckWorkReconciler(ctx context.Context) {
	var requeuedEmbeddings, requeuedResumes, requeuedApplications int

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "requeue pending embeddings",
			Run: func(ctx context.Context) error {
				requeued, customErr := cm.jobService.RequeuePendingEmbeddings(ctx)
				requeuedEmbeddings = requeued

				if customErr != nil {
					return customErr
				}

				return nil
			},
		},
		utils.ConcurrentTask{
			Name: "requeue pending applications",
			Run: func(ctx context.Context) error {
				requeued, customErr := cm.applicationService.RequeueStaleApplications(ctx)
				requeuedApplications = requeued

				if customErr != nil {
					return customErr
				}

				return nil
			},
		},
		utils.ConcurrentTask{
			Name: "requeue pending resumes",
			Run: func(ctx context.Context) error {
				requeued, customErr := cm.candidateService.RequeuePendingResumes(ctx)
				requeuedResumes = requeued

				if customErr != nil {
					return customErr
				}

				return nil
			},
		},
	)
	if err != nil {
		utils.LogError(ctx, err, "cron_job", config.GetStuckWorkReconcilerLock())
	}

	if requeuedEmbeddings+requeuedResumes+requeuedApplications > 0 {
		utils.LogInfo(
			ctx,
			"stuck work reconciled",
			"requeued_embeddings", requeuedEmbeddings,
			"requeued_resumes", requeuedResumes,
			"requeued_applications", requeuedApplications,
		)
	}
}
