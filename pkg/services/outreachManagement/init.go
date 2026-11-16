package outreachManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	enrichmentRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/enrichmentManagement"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	matchingRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/matchingManagement"
	outreachRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement/linkedinProviders"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type OutreachServiceDependencies struct {
	ChannelAccountRepo *outreachRepo.ChannelAccountRepo
	TemplateRepo       *outreachRepo.TemplateRepo
	CampaignRepo       *outreachRepo.CampaignRepo
	MessageRepo        *outreachRepo.MessageRepo
	EventRepo          *outreachRepo.EventRepo
	SuppressionRepo    *outreachRepo.SuppressionRepo
	CandidateRepo      *candidateRepo.CandidateRepo
	ResumeRepo         *candidateRepo.ResumeRepo
	CompanyContactRepo *enrichmentRepo.CompanyContactRepo
	JobPostingRepo     *jobRepo.JobPostingRepo
	JobMatchRepo       *matchingRepo.JobMatchRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
	CredentialCipher   *utils.CredentialCipher
	EmailUtil          *utils.EmailUtil
	LinkedinProvider   linkedinProviders.Provider
}

type OutreachService struct {
	channelAccountRepo *outreachRepo.ChannelAccountRepo
	templateRepo       *outreachRepo.TemplateRepo
	campaignRepo       *outreachRepo.CampaignRepo
	messageRepo        *outreachRepo.MessageRepo
	eventRepo          *outreachRepo.EventRepo
	suppressionRepo    *outreachRepo.SuppressionRepo
	candidateRepo      *candidateRepo.CandidateRepo
	resumeRepo         *candidateRepo.ResumeRepo
	companyContactRepo *enrichmentRepo.CompanyContactRepo
	jobPostingRepo     *jobRepo.JobPostingRepo
	jobMatchRepo       *matchingRepo.JobMatchRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
	credentialCipher   *utils.CredentialCipher
	emailUtil          *utils.EmailUtil
	linkedinProvider   linkedinProviders.Provider
}

func NewOutreachService(dependencies OutreachServiceDependencies) (*OutreachService, error) {
	if dependencies.ChannelAccountRepo == nil ||
		dependencies.TemplateRepo == nil ||
		dependencies.CampaignRepo == nil ||
		dependencies.MessageRepo == nil ||
		dependencies.EventRepo == nil ||
		dependencies.SuppressionRepo == nil ||
		dependencies.CandidateRepo == nil ||
		dependencies.ResumeRepo == nil ||
		dependencies.CompanyContactRepo == nil ||
		dependencies.JobPostingRepo == nil ||
		dependencies.JobMatchRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.EmailUtil == nil ||
		dependencies.LinkedinProvider == nil {
		return nil, fmt.Errorf("new outreach service: %w", config.ErrMissingDependency)
	}

	return &OutreachService{
		channelAccountRepo: dependencies.ChannelAccountRepo,
		templateRepo:       dependencies.TemplateRepo,
		campaignRepo:       dependencies.CampaignRepo,
		messageRepo:        dependencies.MessageRepo,
		eventRepo:          dependencies.EventRepo,
		suppressionRepo:    dependencies.SuppressionRepo,
		candidateRepo:      dependencies.CandidateRepo,
		resumeRepo:         dependencies.ResumeRepo,
		companyContactRepo: dependencies.CompanyContactRepo,
		jobPostingRepo:     dependencies.JobPostingRepo,
		jobMatchRepo:       dependencies.JobMatchRepo,
		transactionManager: dependencies.TransactionManager,
		workerClient:       dependencies.WorkerClient,
		credentialCipher:   dependencies.CredentialCipher,
		emailUtil:          dependencies.EmailUtil,
		linkedinProvider:   dependencies.LinkedinProvider,
	}, nil
}
