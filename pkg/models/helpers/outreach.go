package helpers

type ChannelType string

const ChannelTypeEmail ChannelType = "EMAIL"

type ChannelAccountStatus string

const (
	ChannelAccountStatusUnverified ChannelAccountStatus = "UNVERIFIED"
	ChannelAccountStatusActive     ChannelAccountStatus = "ACTIVE"
	ChannelAccountStatusFailed     ChannelAccountStatus = "FAILED"
)

type MailSecurity string

const (
	MailSecurityTLS      MailSecurity = "TLS"
	MailSecurityStartTLS MailSecurity = "STARTTLS"
	MailSecurityNone     MailSecurity = "NONE"
)

type OutreachChannel string

const (
	OutreachChannelEmail           OutreachChannel = "EMAIL"
	OutreachChannelLinkedinConnect OutreachChannel = "LINKEDIN_CONNECT"
	OutreachChannelLinkedinDM      OutreachChannel = "LINKEDIN_DM"
)

type TemplateStatus string

const (
	TemplateStatusDraft    TemplateStatus = "DRAFT"
	TemplateStatusApproved TemplateStatus = "APPROVED"
	TemplateStatusArchived TemplateStatus = "ARCHIVED"
)

type CampaignStatus string

const (
	CampaignStatusActive   CampaignStatus = "ACTIVE"
	CampaignStatusPaused   CampaignStatus = "PAUSED"
	CampaignStatusArchived CampaignStatus = "ARCHIVED"
)

type OutreachStatus string

const (
	OutreachStatusPending   OutreachStatus = "PENDING"
	OutreachStatusSending   OutreachStatus = "SENDING"
	OutreachStatusSent      OutreachStatus = "SENT"
	OutreachStatusReplied   OutreachStatus = "REPLIED"
	OutreachStatusFailed    OutreachStatus = "FAILED"
	OutreachStatusBounced   OutreachStatus = "BOUNCED"
	OutreachStatusPaused    OutreachStatus = "PAUSED"
	OutreachStatusCancelled OutreachStatus = "CANCELLED"
)

type OutreachEventType string

const (
	OutreachEventSent         OutreachEventType = "SENT"
	OutreachEventOpened       OutreachEventType = "OPENED"
	OutreachEventClicked      OutreachEventType = "CLICKED"
	OutreachEventBounced      OutreachEventType = "BOUNCED"
	OutreachEventReplied      OutreachEventType = "REPLIED"
	OutreachEventAutoReplied  OutreachEventType = "AUTO_REPLIED"
	OutreachEventUnsubscribed OutreachEventType = "UNSUBSCRIBED"
)

type SuppressionType string

const (
	SuppressionTypeEmail  SuppressionType = "EMAIL"
	SuppressionTypeDomain SuppressionType = "DOMAIN"
)

type SuppressionReason string

const (
	SuppressionReasonBounced      SuppressionReason = "BOUNCED"
	SuppressionReasonUnsubscribed SuppressionReason = "UNSUBSCRIBED"
	SuppressionReasonManual       SuppressionReason = "MANUAL"
)
