package authManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/authManagement"
)

type AuthHandler struct {
	authService *authService.AuthService
}

func NewAuthHandler(service *authService.AuthService) (*AuthHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new auth handler: %w", config.ErrMissingDependency)
	}

	return &AuthHandler{authService: service}, nil
}
