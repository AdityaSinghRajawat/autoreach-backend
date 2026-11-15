package outreachManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type Message struct {
	ID                string                       `db:"id"                  json:"outreachId"        validate:"required,uuid"`
	CandidateID       string                       `db:"candidate_id"        json:"candidateId"       validate:"required,uuid"`
	ContactID         string                       `db:"contact_id"          json:"contactId"         validate:"required,uuid"`
	JobID             string                       `db:"job_id"              json:"jobId"             validate:"required,uuid"`
	CampaignID        *string                      `db:"campaign_id"         json:"campaignId"        validate:"omitnil,uuid"`
	ChannelAccountID  *string                      `db:"channel_account_id"  json:"channelAccountId"  validate:"omitnil,uuid"`
	TemplateID        *string                      `db:"template_id"         json:"templateId"        validate:"omitnil,uuid"`
	StepNumber        int                          `db:"step_number"         json:"stepNumber"        validate:"gt=0"`
	Channel           modelHelpers.OutreachChannel `db:"channel"             json:"channel"           validate:"required,oneof=EMAIL LINKEDIN_CONNECT LINKEDIN_DM"`
	Status            modelHelpers.OutreachStatus  `db:"status"              json:"status"            validate:"required,oneof=PENDING SENDING SENT REPLIED FAILED BOUNCED PAUSED CANCELLED"`
	Subject           *string                      `db:"subject"             json:"subject"           validate:"omitnil,max=500"`
	BodyContent       string                       `db:"body_content"        json:"bodyContent"       validate:"required"`
	LinkedinNote      *string                      `db:"linkedin_note"       json:"linkedinNote"      validate:"omitnil,max=300"`
	ScheduledAt       time.Time                    `db:"scheduled_at"        json:"scheduledAt"       validate:"required"`
	ProviderMessageID *string                      `db:"provider_message_id" json:"-"                 validate:"omitnil,max=255"`
	InReplyTo         *string                      `db:"in_reply_to"         json:"-"                 validate:"omitnil,max=255"`
	TrackingToken     *string                      `db:"tracking_token"      json:"-"                 validate:"omitnil,max=64"`
	AttemptCount      int                          `db:"attempt_count"       json:"attemptCount"      validate:"gte=0"`
	SentAt            *time.Time                   `db:"sent_at"             json:"sentAt"`
	RepliedAt         *time.Time                   `db:"replied_at"          json:"repliedAt"`
	FailureReason     *string                      `db:"failure_reason"      json:"failureReason"`
	CreatedAt         time.Time                    `db:"created_at"          json:"createdAt"         validate:"required"`
	UpdatedAt         time.Time                    `db:"updated_at"          json:"updatedAt"         validate:"required"`
}

func NewMessage() *Message {
	now := utils.GetCurrTimeStamp()

	return &Message{
		ID:          uuid.NewString(),
		StepNumber:  1,
		Channel:     modelHelpers.OutreachChannelEmail,
		Status:      modelHelpers.OutreachStatusPending,
		ScheduledAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (m *Message) UpdateTimestamp() {
	m.UpdatedAt = utils.GetCurrTimeStamp()
}

func (m *Message) Validate() error {
	if err := validations.Validate(m); err != nil {
		return fmt.Errorf("outreach message validation failed: %w", err)
	}

	return nil
}

func (m *Message) IsCancellable() bool {
	return m.Status == modelHelpers.OutreachStatusPending ||
		m.Status == modelHelpers.OutreachStatusPaused
}

func (m *Message) Reschedule(at time.Time) {
	m.Status = modelHelpers.OutreachStatusPending
	m.ScheduledAt = at
	m.SentAt = nil
}

func (m *Message) MarkSent() {
	m.Status = modelHelpers.OutreachStatusSent
	m.FailureReason = nil
}

func (m *Message) Close(status modelHelpers.OutreachStatus, reason string) {
	m.Status = status
	m.FailureReason = &reason
}
