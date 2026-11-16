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

	ws.registerWorkerHandlers()

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
		AuthService:        ws.dependencies.Services.Auth,
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

func (ws *WorkerServer) registerWorkerHandlers() {
	services := ws.dependencies.Services
	handlers := workerConfigs.WorkerHandlerRegistry{
		workerConfigs.TaskEmbedJob:           workerTasks.EmbedJob(services.Job),
		workerConfigs.TaskEmbedJobBatch:      workerTasks.EmbedJobBatch(services.Job),
		workerConfigs.TaskProcessResume:      workerTasks.ProcessResume(services.Candidate),
		workerConfigs.TaskMatchResume:        workerTasks.MatchResume(services.Matching),
		workerConfigs.TaskMatchJob:           workerTasks.MatchJob(services.Matching),
		workerConfigs.TaskEnrichContacts:     workerTasks.EnrichContacts(services.Enrichment),
		workerConfigs.TaskPlanOutreach:       workerTasks.PlanOutreach(services.Outreach),
		workerConfigs.TaskSendColdEmail:      workerTasks.SendColdEmail(services.Outreach),
		workerConfigs.TaskSweepMailbox:       workerTasks.SweepMailbox(services.Outreach),
		workerConfigs.TaskSendLinkedinInvite: workerTasks.SendLinkedinInvite(services.Outreach),
		workerConfigs.TaskSendLinkedinDM:     workerTasks.SendLinkedinDM(services.Outreach),
		workerConfigs.TaskQueueApplications:  workerTasks.QueueApplications(services.Application),
		workerConfigs.TaskPrepareApplication: workerTasks.PrepareApplication(services.Application),
		workerConfigs.TaskPollJobSource:      workerTasks.PollJobSource(services.Job),
		workerConfigs.TaskRefreshContact:     workerTasks.RefreshContact(services.Enrichment),
	}

	for taskID, handler := range handlers {
		ws.handlerRegistry[taskID] = wrapWorkerMiddleware(ws.dependencies.Logger, handler)
	}
}
