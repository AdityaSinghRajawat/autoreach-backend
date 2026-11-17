package userManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	authRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/authManagement"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	outreachRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/outreachManagement"
	userRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/userManagement"
)

type UserServiceDependencies struct {
	UserRepo           *userRepo.UserRepo
	UserDataExportRepo *userRepo.UserDataExportRepo
	RefreshTokenRepo   *authRepo.RefreshTokenRepo
	CandidateRepo      *candidateRepo.CandidateRepo
	ResumeRepo         *candidateRepo.ResumeRepo
	CampaignRepo       *outreachRepo.CampaignRepo
	FileStore          sharedTypes.FileStore
	SystemMailer       sharedTypes.SystemMailer
	TransactionManager db.TransactionManager
}

type UserService struct {
	userRepo           *userRepo.UserRepo
	userDataExportRepo *userRepo.UserDataExportRepo
	refreshTokenRepo   *authRepo.RefreshTokenRepo
	candidateRepo      *candidateRepo.CandidateRepo
	resumeRepo         *candidateRepo.ResumeRepo
	campaignRepo       *outreachRepo.CampaignRepo
	fileStore          sharedTypes.FileStore
	systemMailer       sharedTypes.SystemMailer
	transactionManager db.TransactionManager
}

func NewUserService(dependencies UserServiceDependencies) (*UserService, error) {
	if dependencies.UserRepo == nil ||
		dependencies.UserDataExportRepo == nil ||
		dependencies.RefreshTokenRepo == nil ||
		dependencies.CandidateRepo == nil ||
		dependencies.ResumeRepo == nil ||
		dependencies.CampaignRepo == nil ||
		dependencies.FileStore == nil ||
		dependencies.TransactionManager == nil {
		return nil, fmt.Errorf("new user service: %w", config.ErrMissingDependency)
	}

	return &UserService{
		userRepo:           dependencies.UserRepo,
		userDataExportRepo: dependencies.UserDataExportRepo,
		refreshTokenRepo:   dependencies.RefreshTokenRepo,
		candidateRepo:      dependencies.CandidateRepo,
		resumeRepo:         dependencies.ResumeRepo,
		campaignRepo:       dependencies.CampaignRepo,
		fileStore:          dependencies.FileStore,
		systemMailer:       dependencies.SystemMailer,
		transactionManager: dependencies.TransactionManager,
	}, nil
}
