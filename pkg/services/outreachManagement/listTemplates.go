package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) ListTemplates(
	ctx context.Context,
	candidateID string,
) ([]*outreachModel.Template, *config.CustomError) {
	templates, err := s.templateRepo.ListByCandidate(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list templates: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return templates, nil
}
