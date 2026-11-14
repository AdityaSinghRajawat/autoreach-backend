package enrichmentManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement"
)

type EnrichmentHandler struct {
	enrichmentService *enrichmentService.EnrichmentService
}

func NewEnrichmentHandler(
	service *enrichmentService.EnrichmentService,
) (*EnrichmentHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new enrichment handler: %w", config.ErrMissingDependency)
	}

	return &EnrichmentHandler{enrichmentService: service}, nil
}
