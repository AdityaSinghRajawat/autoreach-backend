package cronJobs

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type CronManagerDependencies struct {
	RedisUtil          *utils.RedisUtil
	JobService         *jobService.JobService
	CandidateService   *candidateService.CandidateService
	OutreachService    *outreachService.OutreachService
	ApplicationService *applicationService.ApplicationService
}

type CronManager struct {
	cronInstance       *cron.Cron
	redisUtil          *utils.RedisUtil
	jobService         *jobService.JobService
	candidateService   *candidateService.CandidateService
	outreachService    *outreachService.OutreachService
	applicationService *applicationService.ApplicationService
}

func NewCronManager(dependencies CronManagerDependencies) (*CronManager, error) {
	if dependencies.RedisUtil == nil ||
		dependencies.JobService == nil ||
		dependencies.CandidateService == nil ||
		dependencies.OutreachService == nil ||
		dependencies.ApplicationService == nil {
		return nil, fmt.Errorf("new cron manager: %w", config.ErrMissingDependency)
	}

	return &CronManager{
		cronInstance:       cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		redisUtil:          dependencies.RedisUtil,
		jobService:         dependencies.JobService,
		candidateService:   dependencies.CandidateService,
		outreachService:    dependencies.OutreachService,
		applicationService: dependencies.ApplicationService,
	}, nil
}

func (cm *CronManager) StartJobs(ctx context.Context) {
	cm.stuckWorkReconcilerJob(ctx)
	cm.outreachDispatcherJob(ctx)
	cm.replySweeperJob(ctx)
	cm.jobSourcePollerJob(ctx)
	cm.cronInstance.Start()
}

func (cm *CronManager) StopJobs() {
	select {
	case <-cm.cronInstance.Stop().Done():
	case <-time.After(config.GetCronStopGracePeriod()):
		utils.LogInfoWithoutCtx("cron stop grace period elapsed with jobs still running")
	}
}

func (cm *CronManager) runLocked(
	ctx context.Context,
	name string,
	ttl time.Duration,
	work func(ctx context.Context),
) {
	lock, acquired, err := cm.redisUtil.AcquireLock(ctx, name, ttl)
	if err != nil {
		utils.LogError(ctx, err, "cron_job", name)

		return
	}

	if !acquired {
		return
	}

	defer func() {
		if err := cm.redisUtil.ReleaseLock(context.WithoutCancel(ctx), lock); err != nil {
			utils.LogError(ctx, err, "cron_job", name)
		}
	}()

	work(ctx)
}
