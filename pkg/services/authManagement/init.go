package authManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	authRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/authManagement"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	userRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type AuthServiceDependencies struct {
	UserRepo           *userRepo.UserRepo
	RefreshTokenRepo   *authRepo.RefreshTokenRepo
	APIClientRepo      *authRepo.APIClientRepo
	CandidateRepo      *candidateRepo.CandidateRepo
	TransactionManager db.TransactionManager
	SystemMailer       sharedTypes.SystemMailer
	// RequireEmailVerification makes signup answer the same for new and registered addresses.
	RequireEmailVerification bool
}

type AuthService struct {
	userRepo            *userRepo.UserRepo
	refreshTokenRepo    *authRepo.RefreshTokenRepo
	apiClientRepo       *authRepo.APIClientRepo
	candidateRepo       *candidateRepo.CandidateRepo
	transactionManager  db.TransactionManager
	decoyPasswordHash   string
	systemMailer        sharedTypes.SystemMailer
	requireVerification bool
}

func NewAuthService(dependencies AuthServiceDependencies) (*AuthService, error) {
	if dependencies.UserRepo == nil ||
		dependencies.RefreshTokenRepo == nil ||
		dependencies.APIClientRepo == nil ||
		dependencies.CandidateRepo == nil ||
		dependencies.TransactionManager == nil ||
		(dependencies.RequireEmailVerification && dependencies.SystemMailer == nil) {
		return nil, fmt.Errorf("new auth service: %w", config.ErrMissingDependency)
	}

	decoyToken, _, err := utils.NewOpaqueToken()
	if err != nil {
		return nil, err
	}

	decoyPasswordHash, err := utils.HashPassword(decoyToken)
	if err != nil {
		return nil, err
	}

	return &AuthService{
		userRepo:            dependencies.UserRepo,
		refreshTokenRepo:    dependencies.RefreshTokenRepo,
		apiClientRepo:       dependencies.APIClientRepo,
		candidateRepo:       dependencies.CandidateRepo,
		transactionManager:  dependencies.TransactionManager,
		decoyPasswordHash:   decoyPasswordHash,
		systemMailer:        dependencies.SystemMailer,
		requireVerification: dependencies.RequireEmailVerification,
	}, nil
}
