package authManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	authRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/authManagement"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

const tokenTypeBearer = "Bearer"

type AuthServiceDependencies struct {
	UserRepo           *authRepo.UserRepo
	RefreshTokenRepo   *authRepo.RefreshTokenRepo
	APIClientRepo      *authRepo.APIClientRepo
	CandidateRepo      *candidateRepo.CandidateRepo
	TransactionManager db.TransactionManager
}

type AuthService struct {
	userRepo           *authRepo.UserRepo
	refreshTokenRepo   *authRepo.RefreshTokenRepo
	apiClientRepo      *authRepo.APIClientRepo
	candidateRepo      *candidateRepo.CandidateRepo
	transactionManager db.TransactionManager
	decoyPasswordHash  string
}

func NewAuthService(dependencies AuthServiceDependencies) (*AuthService, error) {
	if dependencies.UserRepo == nil ||
		dependencies.RefreshTokenRepo == nil ||
		dependencies.APIClientRepo == nil ||
		dependencies.CandidateRepo == nil ||
		dependencies.TransactionManager == nil {
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
		userRepo:           dependencies.UserRepo,
		refreshTokenRepo:   dependencies.RefreshTokenRepo,
		apiClientRepo:      dependencies.APIClientRepo,
		candidateRepo:      dependencies.CandidateRepo,
		transactionManager: dependencies.TransactionManager,
		decoyPasswordHash:  decoyPasswordHash,
	}, nil
}
