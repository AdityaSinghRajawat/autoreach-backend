package outreachManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
)

type OutreachHandler struct {
	outreachService *outreachService.OutreachService
}

func NewOutreachHandler(service *outreachService.OutreachService) (*OutreachHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new outreach handler: %w", config.ErrMissingDependency)
	}

	return &OutreachHandler{outreachService: service}, nil
}
