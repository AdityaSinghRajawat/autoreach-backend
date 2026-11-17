package userManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	userService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/userManagement"
)

type UserHandler struct {
	userService *userService.UserService
}

func NewUserHandler(service *userService.UserService) (*UserHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new user handler: %w", config.ErrMissingDependency)
	}

	return &UserHandler{userService: service}, nil
}
