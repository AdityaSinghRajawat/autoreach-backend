package config

import (
	"log"
	"strings"
	"time"
)

const (
	defaultEmailHeloName     = "localhost"
	mailboxInbox             = "INBOX"
	bounceSenderPrefixes     = "mailer-daemon,postmaster"
	maxInboundEmailSize      = 1 << 20
	emailAlternativeBoundary = "autoreach-alternative"
)

type emailConfig struct {
	allowInsecureConnections bool
	timeout                  time.Duration
	heloName                 string
}

var emailConfigI emailConfig

func loadEmailConfig() {
	allowInsecure := getEnvBoolWithDefault("OUTREACH_ALLOW_INSECURE_MAIL", false)
	if allowInsecure && IsProduction() {
		log.Fatalf("OUTREACH_ALLOW_INSECURE_MAIL must not be enabled in production")
	}

	emailConfigI = emailConfig{
		allowInsecureConnections: allowInsecure,
		timeout: getEnvDurationWithDefault(
			"OUTREACH_MAIL_TIMEOUT",
			20*time.Second,
		),
		heloName: getEnvWithDefault(
			"OUTREACH_MAIL_HELO_NAME",
			defaultEmailHeloName,
		),
	}
}

func IsInsecureMailAllowed() bool         { return emailConfigI.allowInsecureConnections }
func GetMailTimeout() time.Duration       { return emailConfigI.timeout }
func GetMailHeloName() string             { return emailConfigI.heloName }
func GetMailboxInbox() string             { return mailboxInbox }
func GetBounceSenderPrefixes() []string   { return strings.Split(bounceSenderPrefixes, ",") }
func GetMaxInboundMessageSize() int64     { return maxInboundEmailSize }
func GetEmailAlternativeBoundary() string { return emailAlternativeBoundary }
