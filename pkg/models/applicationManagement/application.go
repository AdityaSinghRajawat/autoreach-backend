package applicationManagement

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type Application struct {
	ID                   string                         `db:"id"                    json:"applicationId"        validate:"required,uuid"`
	CandidateID          string                         `db:"candidate_id"          json:"candidateId"          validate:"required,uuid"`
	JobID                string                         `db:"job_id"                json:"jobId"                validate:"required,uuid"`
	ResumeID             *string                        `db:"resume_id"             json:"resumeId"             validate:"omitnil,uuid"`
	MatchScore           float64                        `db:"match_score"           json:"matchScore"           validate:"gte=0,lte=1"`
	Status               modelHelpers.ApplicationStatus `db:"status"                json:"status"               validate:"required,oneof=QUEUED IN_PROGRESS SUBMITTED ACTION_REQUIRED FAILED INTERVIEWING REJECTED"`
	ATSType              string                         `db:"ats_type"              json:"atsType"              validate:"required,oneof=GREENHOUSE LEVER WORKDAY OTHER"`
	ApplyURL             string                         `db:"apply_url"             json:"applyUrl"             validate:"required,max=1000"`
	PreparedAnswers      json.RawMessage                `db:"prepared_answers"      json:"preparedAnswers"      validate:"required"`
	MissingAnswers       json.RawMessage                `db:"missing_answers"       json:"missingAnswers"       validate:"required"`
	ExecutionLogs        json.RawMessage                `db:"execution_logs"        json:"executionLogs"        validate:"required"`
	ExternalConfirmation *string                        `db:"external_confirmation" json:"externalConfirmation" validate:"omitnil,max=255"`
	AttemptCount         int                            `db:"attempt_count"         json:"attemptCount"         validate:"gte=0"`
	FailureReason        *string                        `db:"failure_reason"        json:"failureReason"`
	SubmittedAt          *time.Time                     `db:"submitted_at"          json:"submittedAt"`
	CreatedAt            time.Time                      `db:"created_at"            json:"createdAt"            validate:"required"`
	UpdatedAt            time.Time                      `db:"updated_at"            json:"updatedAt"            validate:"required"`
}

func NewApplication() *Application {
	now := utils.GetCurrTimeStamp()

	return &Application{
		ID:              uuid.NewString(),
		Status:          modelHelpers.ApplicationStatusQueued,
		PreparedAnswers: json.RawMessage("[]"),
		MissingAnswers:  json.RawMessage("[]"),
		ExecutionLogs:   json.RawMessage("[]"),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (a *Application) UpdateTimestamp() {
	a.UpdatedAt = utils.GetCurrTimeStamp()
}

func (a *Application) Validate() error {
	if err := validations.Validate(a); err != nil {
		return fmt.Errorf("application validation failed: %w", err)
	}

	return nil
}

func (a *Application) IsPending() bool {
	return a.Status == modelHelpers.ApplicationStatusQueued ||
		a.Status == modelHelpers.ApplicationStatusInProgress
}

func (a *Application) MarkSubmitted(confirmation *string) {
	now := utils.GetCurrTimeStamp()
	a.Status = modelHelpers.ApplicationStatusSubmitted
	a.SubmittedAt = &now
	a.ExternalConfirmation = confirmation
	a.FailureReason = nil
}

func (a *Application) MarkFailed(reason string) {
	a.Status = modelHelpers.ApplicationStatusFailed
	a.FailureReason = &reason
}
