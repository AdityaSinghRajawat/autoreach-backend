package matchingManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
)

type MatchingHandler struct {
	matchingService *matchingService.MatchingService
}

func NewMatchingHandler(service *matchingService.MatchingService) (*MatchingHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new matching handler: %w", config.ErrMissingDependency)
	}

	return &MatchingHandler{matchingService: service}, nil
}
