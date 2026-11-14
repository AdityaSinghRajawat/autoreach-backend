package candidateManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
)

type CandidateHandler struct {
	candidateService *candidateService.CandidateService
}

func NewCandidateHandler(service *candidateService.CandidateService) (*CandidateHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new candidate handler: %w", config.ErrMissingDependency)
	}

	return &CandidateHandler{candidateService: service}, nil
}
