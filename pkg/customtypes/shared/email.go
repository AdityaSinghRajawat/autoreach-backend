package shared

import "time"

type EmailServer struct {
	Host     string
	Port     int
	Security string
}

type EmailAccount struct {
	Address  string
	Username string
	Password string
	SMTP     EmailServer
	IMAP     EmailServer
}

type MailboxCursor struct {
	UIDValidity uint32
	LastUID     uint32
}

type OutgoingEmail struct {
	FromAddress         string
	FromName            string
	ToAddress           string
	ToName              string
	Subject             string
	Body                string
	HTMLBody            string
	MessageID           string
	InReplyTo           string
	Date                time.Time
	ListUnsubscribe     string
	OneClickUnsubscribe bool
}

type InboundEmail struct {
	UID                  uint32
	MessageID            string
	FromAddress          string
	ReferencedIDs        []string
	IsAutoReply          bool
	IsBounce             bool
	IsUnsubscribeRequest bool
	BouncedMessageIDs    []string
	BouncedAddresses     []string
}

// AccountEmail is one of AutoReach's own account emails, rendered as plain text and HTML.
type AccountEmail struct {
	Subject string
	Text    string
	HTML    string
}
