package outreachManagement

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) saveTemplate(
	ctx context.Context,
	template *outreachModel.Template,
	operation string,
) (*outreachModel.Template, *config.CustomError) {
	if details := templateContentIssues(template); len(details) > 0 {
		return nil, config.NewValidationError(
			fmt.Errorf("%s: %w", operation, config.ErrUnknownPlaceholder),
			details,
		)
	}

	template.UpdateTimestamp()

	if err := template.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeBadRequest(),
		)
	}

	saved, err := s.templateRepo.Upsert(ctx, template)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return saved, nil
}

func templateContentIssues(template *outreachModel.Template) []config.ErrorDetail {
	details := []config.ErrorDetail{}

	if template.Channel == modelHelpers.OutreachChannelEmail && template.Subject == nil {
		details = append(
			details,
			config.ErrorDetail{Field: "subject", Issue: config.ErrSubjectRequired.Error()},
		)
	}

	known := config.GetTemplatePlaceholders()
	fields := map[string]string{"body": template.Body}

	if template.Subject != nil {
		fields["subject"] = *template.Subject
	}

	for field, text := range fields {
		if unknown := utils.UnknownPlaceholders(text, known); len(unknown) > 0 {
			details = append(details, config.ErrorDetail{
				Field: field,
				Issue: config.ErrUnknownPlaceholder.Error() + ": " + strings.Join(unknown, ", "),
			})
		}
	}

	return details
}
