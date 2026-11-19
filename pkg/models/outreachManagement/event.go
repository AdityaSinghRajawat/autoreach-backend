package outreachManagement

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type Event struct {
	ID         string                         `db:"id"          validate:"required,uuid"`
	MessageID  string                         `db:"message_id"  validate:"required,uuid"`
	EventType  modelHelpers.OutreachEventType `db:"event_type"  validate:"required,oneof=SENT OPENED CLICKED BOUNCED REPLIED AUTO_REPLIED UNSUBSCRIBED ACCEPTED"`
	Metadata   json.RawMessage                `db:"metadata"    validate:"required,validate_json_object"`
	OccurredAt time.Time                      `db:"occurred_at" validate:"required"`
	CreatedAt  time.Time                      `db:"created_at"  validate:"required"`
}

func NewEvent(messageID string, eventType modelHelpers.OutreachEventType) *Event {
	now := utils.GetCurrTimeStamp()

	return &Event{
		ID:         utils.GenerateUUID(),
		MessageID:  messageID,
		EventType:  eventType,
		Metadata:   json.RawMessage(config.GetEmptyJSONObject()),
		OccurredAt: now,
		CreatedAt:  now,
	}
}

func (e *Event) Validate() error {
	if err := validations.Validate(e); err != nil {
		return fmt.Errorf("outreach event validation failed: %w", err)
	}

	return nil
}
