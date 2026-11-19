package utils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

type EmailUtil struct {
	timeout  time.Duration
	heloName string
}

func NewEmailUtil() *EmailUtil {
	return &EmailUtil{timeout: config.GetMailTimeout(), heloName: config.GetMailHeloName()}
}

func (u *EmailUtil) VerifyEmailAccount(
	ctx context.Context,
	account sharedTypes.EmailAccount,
) (sharedTypes.MailboxCursor, error) {
	smtpClient, err := u.openSMTP(ctx, account)
	if err != nil {
		return sharedTypes.MailboxCursor{}, err
	}

	if err := smtpClient.Quit(); err != nil {
		return sharedTypes.MailboxCursor{}, fmt.Errorf("smtp quit: %w", err)
	}

	imapClient, err := u.openIMAP(ctx, account)
	if err != nil {
		return sharedTypes.MailboxCursor{}, err
	}
	defer func() { _ = imapClient.Close() }()

	selected, err := imapClient.Select(config.GetMailboxInbox(), &imap.SelectOptions{ReadOnly: true}).
		Wait()
	if err != nil {
		return sharedTypes.MailboxCursor{}, fmt.Errorf("imap select: %w", err)
	}

	cursor := sharedTypes.MailboxCursor{UIDValidity: selected.UIDValidity}
	if selected.UIDNext > 0 {
		cursor.LastUID = uint32(selected.UIDNext) - 1
	}

	if err := imapClient.Logout().Wait(); err != nil {
		return sharedTypes.MailboxCursor{}, fmt.Errorf("imap logout: %w", err)
	}

	return cursor, nil
}

func (u *EmailUtil) SendEmail(
	ctx context.Context,
	account sharedTypes.EmailAccount,
	email sharedTypes.OutgoingEmail,
) error {
	raw, err := BuildEmailMessage(email)
	if err != nil {
		return fmt.Errorf("compose message: %w", err)
	}

	client, err := u.openSMTP(ctx, account)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if err := client.Mail(email.FromAddress); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(email.ToAddress); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := writer.Write(raw); err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	// Only the DATA reply decides the outcome; a failed QUIT after acceptance does not matter.
	if err := writer.Close(); err != nil {
		var reply *textproto.Error
		if errors.As(err, &reply) {
			return fmt.Errorf("smtp data: %w", err)
		}

		return fmt.Errorf("smtp data: %w: %w", config.ErrSendOutcomeUnknown, err)
	}

	_ = client.Quit()

	return nil
}

func (u *EmailUtil) openSMTP(
	ctx context.Context,
	account sharedTypes.EmailAccount,
) (*smtp.Client, error) {
	server := account.SMTP

	conn, err := u.dialEmailServer(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("smtp connect: %w: %w", config.ErrMailServerUnreachable, err)
	}

	client, err := smtp.NewClient(conn, server.Host)
	if err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("smtp greeting: %w: %w", config.ErrMailServerUnreachable, err)
	}

	if err := u.authenticateSMTP(client, account); err != nil {
		_ = client.Close()

		return nil, err
	}

	return client, nil
}

func (u *EmailUtil) authenticateSMTP(client *smtp.Client, account sharedTypes.EmailAccount) error {
	if err := client.Hello(u.heloName); err != nil {
		return fmt.Errorf("smtp hello: %w: %w", config.ErrMailServerUnreachable, err)
	}

	if account.SMTP.Security == string(modelHelpers.MailSecurityStartTLS) {
		if err := client.StartTLS(emailTLSConfig(account.SMTP.Host)); err != nil {
			return fmt.Errorf("smtp starttls: %w: %w", config.ErrMailSecureConnection, err)
		}
	}

	if supported, _ := client.Extension("AUTH"); !supported {
		return fmt.Errorf("smtp auth: %w", config.ErrMailAuthUnsupported)
	}

	if err := client.Auth(
		smtp.PlainAuth("", account.Username, account.Password, account.SMTP.Host),
	); err != nil {
		return fmt.Errorf("smtp auth: %w: %w", config.ErrMailLoginRejected, err)
	}

	return nil
}

func (u *EmailUtil) openIMAP(
	ctx context.Context,
	account sharedTypes.EmailAccount,
) (*imapclient.Client, error) {
	server := account.IMAP

	conn, err := u.dialEmailServer(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("imap connect: %w: %w", config.ErrMailServerUnreachable, err)
	}

	var client *imapclient.Client

	if server.Security == string(modelHelpers.MailSecurityStartTLS) {
		client, err = imapclient.NewStartTLS(
			conn,
			&imapclient.Options{TLSConfig: emailTLSConfig(server.Host)},
		)
		if err != nil {
			_ = conn.Close()

			return nil, fmt.Errorf("imap starttls: %w: %w", config.ErrMailSecureConnection, err)
		}
	} else {
		client = imapclient.New(conn, nil)
	}

	if err := client.Login(account.Username, account.Password).Wait(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("imap login: %w: %w", config.ErrMailLoginRejected, err)
	}

	return client, nil
}

// dialEmailServer opens the connection itself so one deadline bounds the whole mail session.
func (u *EmailUtil) dialEmailServer(
	ctx context.Context,
	server sharedTypes.EmailServer,
) (net.Conn, error) {
	deadline := time.Now().Add(u.timeout)
	if ctxDeadline, found := ctx.Deadline(); found && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	dialer := net.Dialer{Deadline: deadline}
	if !config.IsPrivateMailHostAllowed() {
		dialer.Control = publicAddressOnly
	}

	address := net.JoinHostPort(server.Host, strconv.Itoa(server.Port))

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	if server.Security == string(modelHelpers.MailSecurityTLS) {
		conn = tls.Client(conn, emailTLSConfig(server.Host))
	}

	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()

		return nil, err
	}

	return conn, nil
}

func emailTLSConfig(host string) *tls.Config {
	return &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
}

// A changed UIDVALIDITY invalidates old UIDs, so the cursor restarts at the newest message.
func (u *EmailUtil) FetchInboxSince(
	ctx context.Context,
	account sharedTypes.EmailAccount,
	cursor sharedTypes.MailboxCursor,
) ([]sharedTypes.InboundEmail, sharedTypes.MailboxCursor, error) {
	client, err := u.openIMAP(ctx, account)
	if err != nil {
		return nil, cursor, err
	}
	defer func() { _ = client.Close() }()

	selected, err := client.Select(config.GetMailboxInbox(), &imap.SelectOptions{ReadOnly: true}).
		Wait()
	if err != nil {
		return nil, cursor, fmt.Errorf("imap select: %w", err)
	}

	latest := sharedTypes.MailboxCursor{UIDValidity: selected.UIDValidity}
	if selected.UIDNext > 0 {
		latest.LastUID = uint32(selected.UIDNext) - 1
	}

	if selected.UIDValidity != cursor.UIDValidity || latest.LastUID <= cursor.LastUID {
		return nil, latest, client.Logout().Wait()
	}

	newMail := imap.UIDSet{{Start: imap.UID(cursor.LastUID + 1), Stop: imap.UID(latest.LastUID)}}

	sizes, err := client.Fetch(newMail, &imap.FetchOptions{UID: true, RFC822Size: true}).Collect()
	if err != nil {
		return nil, cursor, fmt.Errorf("imap fetch sizes: %w", err)
	}

	readable := imap.UIDSet{}

	for _, message := range sizes {
		if message.RFC822Size <= config.GetMaxInboundMessageSize() {
			readable.AddNum(message.UID)
		}
	}

	if len(readable) == 0 {
		return nil, latest, client.Logout().Wait()
	}

	section := &imap.FetchItemBodySection{Peek: true}

	fetched, err := client.Fetch(
		readable,
		&imap.FetchOptions{UID: true, BodySection: []*imap.FetchItemBodySection{section}},
	).Collect()
	if err != nil {
		return nil, cursor, fmt.Errorf("imap fetch: %w", err)
	}

	inbound := make([]sharedTypes.InboundEmail, 0, len(fetched))

	for _, message := range fetched {
		parsed, err := ParseInboundEmail(uint32(message.UID), message.FindBodySection(section))
		if err != nil {
			continue
		}

		inbound = append(inbound, parsed)
	}

	return inbound, latest, client.Logout().Wait()
}

// A category only: the server's reply would reveal what listens at the address.
func MailFailureReason(err error) string {
	for _, category := range []error{
		config.ErrMailAuthUnsupported,
		config.ErrMailLoginRejected,
		config.ErrMailSecureConnection,
		config.ErrMailServerUnreachable,
	} {
		if errors.Is(err, category) {
			return category.Error()
		}
	}

	return config.ErrMailboxCheckFailed.Error()
}

type MXResolver interface {
	LookupMX(ctx context.Context, domain string) ([]*net.MX, error)
}

type EmailVerifier struct {
	resolver MXResolver
}

func NewEmailVerifier(resolver MXResolver) *EmailVerifier {
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	return &EmailVerifier{resolver: resolver}
}

func (v *EmailVerifier) Verify(
	ctx context.Context,
	email string,
	providerVerified bool,
) modelHelpers.EmailVerificationStatus {
	localPart, domain, valid := splitEmail(email)
	if !valid || config.IsDisposableDomain(domain) {
		return modelHelpers.EmailStatusInvalid
	}

	lookupCtx, cancel := context.WithTimeout(ctx, config.GetMXLookupTimeout())
	defer cancel()

	records, err := v.resolver.LookupMX(lookupCtx, domain)
	if err != nil || len(records) == 0 {
		return modelHelpers.EmailStatusInvalid
	}

	if config.IsRoleLocalPart(localPart) || !providerVerified {
		return modelHelpers.EmailStatusRisky
	}

	return modelHelpers.EmailStatusVerified
}

func splitEmail(email string) (localPart, domain string, valid bool) {
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", "", false
	}

	localPart, domain, found := strings.Cut(email, "@")
	if !found || localPart == "" || !strings.Contains(domain, ".") {
		return "", "", false
	}

	return localPart, strings.ToLower(domain), true
}

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
