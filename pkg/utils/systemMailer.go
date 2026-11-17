package utils

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

// Sends from AutoReach's own address, not a candidate's mailbox.
type SystemMailer struct {
	emailUtil *EmailUtil
}

var _ sharedTypes.SystemMailer = (*SystemMailer)(nil)

func NewSystemMailer(emailUtil *EmailUtil) (*SystemMailer, error) {
	if emailUtil == nil {
		return nil, fmt.Errorf("new system mailer: %w", config.ErrMissingDependency)
	}

	return &SystemMailer{emailUtil: emailUtil}, nil
}

func (m *SystemMailer) Send(
	ctx context.Context,
	toAddress string,
	email sharedTypes.AccountEmail,
) error {
	if !config.IsSystemMailConfigured() {
		return fmt.Errorf("send system mail: %w", config.ErrSystemMailNotConfigured)
	}

	server := sharedTypes.EmailServer{
		Host:     config.GetSystemSMTPHost(),
		Port:     config.GetSystemSMTPPort(),
		Security: config.GetSystemSMTPSecurity(),
	}

	return m.emailUtil.SendEmail(ctx, sharedTypes.EmailAccount{
		Address:  config.GetSystemMailAddress(),
		Username: config.GetSystemSMTPUsername(),
		Password: config.GetSystemSMTPPassword(),
		SMTP:     server,
		IMAP:     server,
	}, sharedTypes.OutgoingEmail{
		FromAddress: config.GetSystemMailAddress(),
		FromName:    config.GetSystemMailName(),
		ToAddress:   toAddress,
		Subject:     email.Subject,
		Body:        email.Text,
		HTMLBody:    email.HTML,
		MessageID:   NewEmailMessageID(config.GetSystemMailAddress()),
		Date:        GetCurrTimeStamp(),
	})
}

// SendAccountEmailInBackground keeps response time from revealing whether an account exists.
func SendAccountEmailInBackground(
	ctx context.Context,
	mailer sharedTypes.SystemMailer,
	userID, address string,
	email sharedTypes.AccountEmail,
) {
	if mailer == nil {
		return
	}

	RunInBackground(ctx, "send account email", func(ctx context.Context) {
		if err := mailer.Send(ctx, address, email); err != nil {
			LogWarn(ctx, "account email not sent", "user_id", userID, "error", err.Error())
		}
	})
}
