package applicationManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
)

type ApplicationHandler struct {
	applicationService *applicationService.ApplicationService
}

func NewApplicationHandler(
	service *applicationService.ApplicationService,
) (*ApplicationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new application handler: %w", config.ErrMissingDependency)
	}

	return &ApplicationHandler{applicationService: service}, nil
}
