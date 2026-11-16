package adminManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	adminService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/adminManagement"
)

type AdminHandler struct {
	adminService *adminService.AdminService
}

func NewAdminHandler(service *adminService.AdminService) (*AdminHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new admin handler: %w", config.ErrMissingDependency)
	}

	return &AdminHandler{adminService: service}, nil
}
