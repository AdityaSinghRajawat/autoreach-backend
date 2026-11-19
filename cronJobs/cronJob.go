package cronJobs

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	authService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/authManagement"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type CronManagerDependencies struct {
	RedisUtil          *utils.RedisUtil
	AuthService        *authService.AuthService
	JobService         *jobService.JobService
	CandidateService   *candidateService.CandidateService
	OutreachService    *outreachService.OutreachService
	ApplicationService *applicationService.ApplicationService
}

type CronManager struct {
	cronInstance       *cron.Cron
	redisUtil          *utils.RedisUtil
	authService        *authService.AuthService
	jobService         *jobService.JobService
	candidateService   *candidateService.CandidateService
	outreachService    *outreachService.OutreachService
	applicationService *applicationService.ApplicationService
}

func NewCronManager(dependencies CronManagerDependencies) (*CronManager, error) {
	if dependencies.RedisUtil == nil ||
		dependencies.AuthService == nil ||
		dependencies.JobService == nil ||
		dependencies.CandidateService == nil ||
		dependencies.OutreachService == nil ||
		dependencies.ApplicationService == nil {
		return nil, fmt.Errorf("new cron manager: %w", config.ErrMissingDependency)
	}

	return &CronManager{
		cronInstance:       cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		redisUtil:          dependencies.RedisUtil,
		authService:        dependencies.AuthService,
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
	cm.dataRetentionJob(ctx)
	cm.cronInstance.Start()
}

func (cm *CronManager) StopJobs() {
	select {
	case <-cm.cronInstance.Stop().Done():
	case <-time.After(config.GetCronStopGracePeriod()):
		utils.LogInfoWithoutCtx("cron stop grace period elapsed with jobs still running")
	}
}

// The claim outlives the run, so other replicas skip ticks inside the interval.
func (cm *CronManager) runOncePerInterval(
	ctx context.Context,
	name string,
	interval time.Duration,
	work func(ctx context.Context),
) {
	claimed, err := cm.redisUtil.SetIfNotExists(
		ctx,
		utils.BuildKey(
			config.GetKeyDelimiter(),
			config.GetKeyPrefix(),
			config.GetLockNamespace(),
			name,
		),
		utils.GenerateUUID(),
		config.GetCronClaimHold(interval),
	)
	if err != nil {
		utils.LogError(ctx, err, "cron_job", name)

		return
	}

	if claimed {
		work(ctx)
	}
}
