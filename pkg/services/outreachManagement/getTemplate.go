package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) GetTemplate(
	ctx context.Context,
	candidateID, templateID string,
) (*outreachModel.Template, *config.CustomError) {
	template, err := s.templateRepo.FindByCandidate(ctx, candidateID, templateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get template: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if template == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get template: %w", config.ErrTemplateNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return template, nil
}
