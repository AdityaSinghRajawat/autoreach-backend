package jobManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
)

type JobHandler struct {
	jobService *jobService.JobService
}

func NewJobHandler(service *jobService.JobService) (*JobHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new job handler: %w", config.ErrMissingDependency)
	}

	return &JobHandler{jobService: service}, nil
}
