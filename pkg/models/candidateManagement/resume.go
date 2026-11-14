package candidateManagement

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type Resume struct {
	ID             string                    `db:"id"              json:"resumeId"       validate:"required,uuid"`
	CandidateID    string                    `db:"candidate_id"    json:"candidateId"    validate:"required,uuid"`
	FileName       string                    `db:"file_name"       json:"fileName"       validate:"required,max=255"`
	StorageKey     string                    `db:"storage_key"     json:"-"              validate:"required,max=500"`
	ContentType    string                    `db:"content_type"    json:"contentType"    validate:"required,oneof=application/pdf application/vnd.openxmlformats-officedocument.wordprocessingml.document"`
	SizeBytes      int64                     `db:"size_bytes"      json:"sizeBytes"      validate:"gt=0"`
	ContentSHA256  string                    `db:"content_sha256"  json:"-"              validate:"required,len=64,hexadecimal"`
	RawText        string                    `db:"raw_text"        json:"-"`
	ParsedMetadata json.RawMessage           `db:"parsed_metadata" json:"parsedMetadata" validate:"required,validate_json_object"`
	Status         modelHelpers.ResumeStatus `db:"parse_status"    json:"status"         validate:"required,oneof=UPLOADED PARSED EMBEDDED FAILED"`
	FailureReason  *string                   `db:"failure_reason"  json:"failureReason"`
	EmbeddingModel *string                   `db:"embedding_model" json:"-"`
	IsActive       bool                      `db:"is_active"       json:"isActive"`
	CreatedAt      time.Time                 `db:"created_at"      json:"createdAt"      validate:"required"`
	UpdatedAt      time.Time                 `db:"updated_at"      json:"updatedAt"      validate:"required"`
}

func NewResume() *Resume {
	now := utils.GetCurrTimeStamp()

	return &Resume{
		ID:             uuid.NewString(),
		ParsedMetadata: json.RawMessage(config.GetEmptyJSONObject()),
		Status:         modelHelpers.ResumeStatusUploaded,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (r *Resume) UpdateTimestamp() {
	r.UpdatedAt = utils.GetCurrTimeStamp()
}

func (r *Resume) Validate() error {
	if err := validations.Validate(r); err != nil {
		return fmt.Errorf("resume validation failed: %w", err)
	}

	return nil
}

func (r *Resume) IsEmbedded() bool {
	return r.Status == modelHelpers.ResumeStatusEmbedded
}

func (r *Resume) MarkParsed(rawText string, parsedMetadata json.RawMessage) {
	r.RawText = rawText
	r.ParsedMetadata = parsedMetadata
	r.Status = modelHelpers.ResumeStatusParsed
	r.FailureReason = nil
}

func (r *Resume) MarkFailed(reason string) {
	r.Status = modelHelpers.ResumeStatusFailed
	r.FailureReason = &reason
}
