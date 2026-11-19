package outreachManagement

import (
	"strings"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type CreateTemplateRequest struct {
	Name    string  `json:"name"    validate:"required,max=255"`
	Channel string  `json:"channel" validate:"required,oneof=EMAIL LINKEDIN_CONNECT LINKEDIN_DM"`
	Subject *string `json:"subject" validate:"omitnil,max=500"`
	Body    string  `json:"body"    validate:"required,max=20000"`
}

type UpdateTemplateRequest struct {
	Name    *string `json:"name"    validate:"omitnil,min=1,max=255"`
	Subject *string `json:"subject" validate:"omitnil,max=500"`
	Body    *string `json:"body"    validate:"omitnil,min=1,max=20000"`
	Status  *string `json:"status"  validate:"omitnil,oneof=DRAFT APPROVED ARCHIVED"`
}

func (r *CreateTemplateRequest) ToTemplateModel(candidateID string) *outreachModel.Template {
	template := outreachModel.NewTemplate()
	template.CandidateID = candidateID
	template.Name = strings.TrimSpace(r.Name)
	template.Channel = modelHelpers.OutreachChannel(r.Channel)
	template.Subject = utils.TrimOptionalStringToNil(r.Subject)
	template.Body = r.Body

	return template
}

func (r *UpdateTemplateRequest) IsEmpty() bool {
	return r.Name == nil && r.Subject == nil && r.Body == nil && r.Status == nil
}

func (r *UpdateTemplateRequest) ChangesContent() bool {
	return r.Subject != nil || r.Body != nil
}

// An edit to the wording of an approved template sends it back to draft for re-approval.
func (r *UpdateTemplateRequest) ToTemplateModel(
	existing *outreachModel.Template,
) *outreachModel.Template {
	template := *existing

	if r.Name != nil {
		template.Name = strings.TrimSpace(*r.Name)
	}

	if r.Subject != nil {
		template.Subject = utils.TrimOptionalStringToNil(r.Subject)
	}

	if r.Body != nil {
		template.Body = *r.Body
	}

	switch {
	case r.Status != nil:
		template.SetStatus(modelHelpers.TemplateStatus(*r.Status))
	case r.ChangesContent() && template.IsApproved():
		template.SetStatus(modelHelpers.TemplateStatusDraft)
	}

	return &template
}
