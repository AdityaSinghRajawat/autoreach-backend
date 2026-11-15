package workerServer

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/cronJobs"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	workerAggregators "github.com/AdityaSinghRajawat/autoreach-backend/worker/aggregators"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerTasks "github.com/AdityaSinghRajawat/autoreach-backend/worker/tasks"
)

type taskHandler interface {
	HandleWorkerMessage(
		ctx context.Context,
		message workerConfigs.Message,
	) *workerConfigs.WorkerMessageResult
}

type WorkerServer struct {
	dependencies       *dependencies.Dependencies
	cronManager        *cronJobs.CronManager
	handlerRegistry    workerConfigs.WorkerHandlerRegistry
	aggregatorRegistry workerConfigs.WorkerAggregatorRegistry
}

func NewWorkerServer(_ context.Context, deps *dependencies.Dependencies) (*WorkerServer, error) {
	if deps == nil || deps.WorkerStrategy == nil {
		return nil, fmt.Errorf("new worker server: %w", config.ErrMissingDependency)
	}

	ws := &WorkerServer{
		dependencies:       deps,
		handlerRegistry:    workerConfigs.WorkerHandlerRegistry{},
		aggregatorRegistry: workerAggregators.NewWorkerAggregatorRegistry(),
	}

	if err := ws.loadCrons(); err != nil {
		return nil, err
	}

	if err := ws.registerWorkerHandlers(); err != nil {
		return nil, err
	}

	if err := workerConfigs.ValidateTopology(
		ws.handlerRegistry,
		ws.aggregatorRegistry,
	); err != nil {
		return nil, fmt.Errorf("validate worker topology: %w", err)
	}

	return ws, nil
}

func (ws *WorkerServer) loadCrons() error {
	cronManager, err := cronJobs.NewCronManager(cronJobs.CronManagerDependencies{
		OutreachService:    ws.dependencies.Services.Outreach,
		ApplicationService: ws.dependencies.Services.Application,
		RedisUtil:          ws.dependencies.RedisUtil,
		JobService:         ws.dependencies.Services.Job,
		CandidateService:   ws.dependencies.Services.Candidate,
	})
	if err != nil {
		return err
	}

	ws.cronManager = cronManager

	return nil
}

func (ws *WorkerServer) registerWorkerHandlers() error {
	services := ws.dependencies.Services

	embedJob, err := workerTasks.NewEmbedJobTaskHandler(services.Job)
	if err != nil {
		return err
	}

	embedJobBatch, err := workerTasks.NewEmbedJobBatchTaskHandler(services.Job)
	if err != nil {
		return err
	}

	processResume, err := workerTasks.NewProcessResumeTaskHandler(services.Candidate)
	if err != nil {
		return err
	}

	matchResume, err := workerTasks.NewMatchResumeTaskHandler(services.Matching)
	if err != nil {
		return err
	}

	matchJob, err := workerTasks.NewMatchJobTaskHandler(services.Matching)
	if err != nil {
		return err
	}

	enrichContacts, err := workerTasks.NewEnrichContactsTaskHandler(services.Enrichment)
	if err != nil {
		return err
	}

	planOutreach, err := workerTasks.NewPlanOutreachTaskHandler(services.Outreach)
	if err != nil {
		return err
	}

	sendColdEmail, err := workerTasks.NewSendColdEmailTaskHandler(services.Outreach)
	if err != nil {
		return err
	}

	sweepMailbox, err := workerTasks.NewSweepMailboxTaskHandler(services.Outreach)
	if err != nil {
		return err
	}

	sendLinkedinInvite, err := workerTasks.NewSendLinkedinInviteTaskHandler(services.Outreach)
	if err != nil {
		return err
	}

	sendLinkedinDM, err := workerTasks.NewSendLinkedinDMTaskHandler(services.Outreach)
	if err != nil {
		return err
	}

	queueApplications, err := workerTasks.NewQueueApplicationsTaskHandler(services.Application)
	if err != nil {
		return err
	}

	prepareApplication, err := workerTasks.NewPrepareApplicationTaskHandler(services.Application)
	if err != nil {
		return err
	}

	ws.register(workerConfigs.TaskEmbedJob, embedJob)
	ws.register(workerConfigs.TaskEmbedJobBatch, embedJobBatch)
	ws.register(workerConfigs.TaskProcessResume, processResume)
	ws.register(workerConfigs.TaskMatchResume, matchResume)
	ws.register(workerConfigs.TaskMatchJob, matchJob)
	ws.register(workerConfigs.TaskEnrichContacts, enrichContacts)
	ws.register(workerConfigs.TaskPlanOutreach, planOutreach)
	ws.register(workerConfigs.TaskSendColdEmail, sendColdEmail)
	ws.register(workerConfigs.TaskSweepMailbox, sweepMailbox)
	ws.register(workerConfigs.TaskSendLinkedinInvite, sendLinkedinInvite)
	ws.register(workerConfigs.TaskSendLinkedinDM, sendLinkedinDM)
	ws.register(workerConfigs.TaskQueueApplications, queueApplications)
	ws.register(workerConfigs.TaskPrepareApplication, prepareApplication)

	return nil
}

func (ws *WorkerServer) register(taskID workerConfigs.WorkerTaskID, handler taskHandler) {
	ws.handlerRegistry[taskID] = wrapWorkerMiddleware(
		ws.dependencies.Logger,
		handler.HandleWorkerMessage,
	)
}
