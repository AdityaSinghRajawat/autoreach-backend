package outreachManagement

import (
	"strings"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type CreateChannelAccountRequest struct {
	EmailAddress string  `json:"emailAddress" validate:"required,email,max=255"`
	DisplayName  *string `json:"displayName"  validate:"omitnil,max=255"`
	Username     string  `json:"username"     validate:"required,max=255"`
	Password     string  `json:"password"     validate:"required,max=1024"`
	SMTPHost     string  `json:"smtpHost"     validate:"required,hostname_rfc1123|ip"`
	SMTPPort     int     `json:"smtpPort"     validate:"required,min=1,max=65535"`
	SMTPSecurity string  `json:"smtpSecurity" validate:"required,oneof=TLS STARTTLS NONE"`
	IMAPHost     string  `json:"imapHost"     validate:"required,hostname_rfc1123|ip"`
	IMAPPort     int     `json:"imapPort"     validate:"required,min=1,max=65535"`
	IMAPSecurity string  `json:"imapSecurity" validate:"required,oneof=TLS STARTTLS NONE"`
	DailyLimit   *int    `json:"dailyLimit"   validate:"omitnil,min=1"`
}

type ChannelAccountCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *CreateChannelAccountRequest) ToChannelAccountModel(
	candidateID string,
) *outreachModel.ChannelAccount {
	account := outreachModel.NewChannelAccount()
	account.CandidateID = candidateID
	account.EmailAddress = strings.ToLower(strings.TrimSpace(r.EmailAddress))
	account.SMTPHost = strings.ToLower(r.SMTPHost)
	account.SMTPPort = r.SMTPPort
	account.SMTPSecurity = modelHelpers.MailSecurity(r.SMTPSecurity)
	account.IMAPHost = strings.ToLower(r.IMAPHost)
	account.IMAPPort = r.IMAPPort
	account.IMAPSecurity = modelHelpers.MailSecurity(r.IMAPSecurity)

	if r.DisplayName != nil {
		if displayName := strings.TrimSpace(*r.DisplayName); displayName != "" {
			account.DisplayName = &displayName
		}
	}

	if r.DailyLimit != nil {
		account.DailyLimit = min(*r.DailyLimit, account.DailyLimit)
	}

	return account
}

func (r *CreateChannelAccountRequest) Credentials() ChannelAccountCredentials {
	return ChannelAccountCredentials{Username: r.Username, Password: r.Password}
}
