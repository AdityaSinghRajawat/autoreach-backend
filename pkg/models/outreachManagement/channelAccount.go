package outreachManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type ChannelAccount struct {
	ID                   string                            `db:"id"                    json:"accountId"      validate:"required,uuid"`
	CandidateID          string                            `db:"candidate_id"          json:"candidateId"    validate:"required,uuid"`
	ChannelType          modelHelpers.ChannelType          `db:"channel_type"          json:"channelType"    validate:"required,oneof=EMAIL"`
	EmailAddress         string                            `db:"identifier"            json:"emailAddress"   validate:"required,email,max=255"`
	DisplayName          *string                           `db:"display_name"          json:"displayName"    validate:"omitnil,max=255"`
	SMTPHost             string                            `db:"smtp_host"             json:"smtpHost"       validate:"required,max=255"`
	SMTPPort             int                               `db:"smtp_port"             json:"smtpPort"       validate:"min=1,max=65535"`
	SMTPSecurity         modelHelpers.MailSecurity         `db:"smtp_security"         json:"smtpSecurity"   validate:"required,oneof=TLS STARTTLS NONE"`
	IMAPHost             string                            `db:"imap_host"             json:"imapHost"       validate:"required,max=255"`
	IMAPPort             int                               `db:"imap_port"             json:"imapPort"       validate:"min=1,max=65535"`
	IMAPSecurity         modelHelpers.MailSecurity         `db:"imap_security"         json:"imapSecurity"   validate:"required,oneof=TLS STARTTLS NONE"`
	EncryptedCredentials []byte                            `db:"encrypted_credentials" json:"-"              validate:"required"`
	KeyVersion           int                               `db:"key_version"           json:"-"              validate:"gt=0"`
	Status               modelHelpers.ChannelAccountStatus `db:"status"                json:"status"         validate:"required,oneof=UNVERIFIED ACTIVE FAILED"`
	DailyLimit           int                               `db:"daily_limit"           json:"dailyLimit"     validate:"gt=0"`
	LastVerifiedAt       *time.Time                        `db:"last_verified_at"      json:"lastVerifiedAt"`
	LastError            *string                           `db:"last_error"            json:"lastError"`
	IMAPUIDValidity      *int64                            `db:"imap_uid_validity"     json:"-"`
	IMAPLastUID          *int64                            `db:"imap_last_uid"         json:"-"`
	CreatedAt            time.Time                         `db:"created_at"            json:"createdAt"      validate:"required"`
	UpdatedAt            time.Time                         `db:"updated_at"            json:"updatedAt"      validate:"required"`
}

func NewChannelAccount() *ChannelAccount {
	now := utils.GetCurrTimeStamp()

	return &ChannelAccount{
		ID:          uuid.NewString(),
		ChannelType: modelHelpers.ChannelTypeEmail,
		Status:      modelHelpers.ChannelAccountStatusUnverified,
		DailyLimit:  config.GetMaxDailyEmailsPerSender(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (a *ChannelAccount) UpdateTimestamp() {
	a.UpdatedAt = utils.GetCurrTimeStamp()
}

func (a *ChannelAccount) Validate() error {
	if err := validations.Validate(a); err != nil {
		return fmt.Errorf("channel account validation failed: %w", err)
	}

	return nil
}

func (a *ChannelAccount) UsesInsecureConnection() bool {
	return a.SMTPSecurity == modelHelpers.MailSecurityNone ||
		a.IMAPSecurity == modelHelpers.MailSecurityNone
}

func (a *ChannelAccount) MarkVerified(cursor sharedTypes.MailboxCursor) {
	now := utils.GetCurrTimeStamp()
	uidValidity := int64(cursor.UIDValidity)
	lastUID := int64(cursor.LastUID)

	a.Status = modelHelpers.ChannelAccountStatusActive
	a.LastVerifiedAt = &now
	a.LastError = nil
	a.IMAPUIDValidity = &uidValidity
	a.IMAPLastUID = &lastUID
}

func (a *ChannelAccount) MarkFailed(reason string) {
	now := utils.GetCurrTimeStamp()

	if limit := config.GetMaxChannelAccountErrorLength(); len(reason) > limit {
		reason = reason[:limit]
	}

	a.Status = modelHelpers.ChannelAccountStatusFailed
	a.LastVerifiedAt = &now
	a.LastError = &reason
}
