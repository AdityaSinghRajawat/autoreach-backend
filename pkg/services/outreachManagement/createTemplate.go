package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) CreateTemplate(
	ctx context.Context,
	template *outreachModel.Template,
) (*outreachModel.Template, *config.CustomError) {
	return s.saveTemplate(ctx, template, "create template")
}
