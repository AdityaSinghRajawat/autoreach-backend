package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) GetTemplate(
	ctx context.Context,
	candidateID, templateID string,
) (*outreachModel.Template, *config.CustomError) {
	template, err := s.templateRepo.FindByCandidate(ctx, candidateID, templateID)

	return config.RequireFound("get template", template, err, config.ErrTemplateNotFound)
}
