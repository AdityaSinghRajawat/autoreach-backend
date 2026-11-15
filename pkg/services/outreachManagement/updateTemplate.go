package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) UpdateTemplate(
	ctx context.Context,
	candidateID, templateID string,
	request outreachTypes.UpdateTemplateRequest,
) (*outreachModel.Template, *config.CustomError) {
	if request.IsEmpty() {
		return nil, config.NewCustomError(
			fmt.Errorf("update template: %w", config.ErrNoFieldsSupplied),
			config.GetErrCodeBadRequest(),
		)
	}

	existing, customErr := s.GetTemplate(ctx, candidateID, templateID)
	if customErr != nil {
		return nil, customErr
	}

	return s.saveTemplate(ctx, request.ToTemplateModel(existing), "update template")
}
