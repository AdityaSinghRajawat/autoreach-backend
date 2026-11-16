package outreachManagement

import (
	"strings"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type CreateSuppressionRequest struct {
	EntryType   string  `json:"entryType"   validate:"required,oneof=EMAIL DOMAIN"`
	Value       string  `json:"value"       validate:"required,max=255,validate_suppression_value"`
	CandidateID *string `json:"candidateId" validate:"omitnil,uuid"`
}

type ListSuppressionsParams struct {
	sharedTypes.PageParams
	Query       string `form:"q"           validate:"omitempty,max=255"`
	CandidateID string `form:"candidateId" validate:"omitempty,uuid"`
}

func (r *CreateSuppressionRequest) ToSuppressionEntryModel() *outreachModel.SuppressionEntry {
	entry := outreachModel.NewSuppressionEntry()
	entry.EntryType = modelHelpers.SuppressionType(r.EntryType)
	entry.Value = strings.ToLower(strings.TrimSpace(r.Value))
	entry.Reason = modelHelpers.SuppressionReasonManual
	entry.CandidateID = r.CandidateID

	return entry
}
