package config

import (
	"log"
	"time"
)

const (
	emailVerificationKeyLabel = "email-verification"
	passwordResetKeyLabel     = "password-reset"
	accountTokenSeparator     = "."
	accountTokenParam         = "token"
	accountTokenParts         = 3
)

type accountEmailConfig struct {
	requireVerification bool
	verificationTTL     time.Duration
	verificationURL     string
	resetTTL            time.Duration
	resetURL            string
	senderAddress       string
	senderName          string
	smtpHost            string
	smtpPort            int
	smtpSecurity        string
	smtpUsername        string
	smtpPassword        string
}

var accountEmailConfigI accountEmailConfig

func loadAccountEmailConfig() {
	accountEmailConfigI = accountEmailConfig{
		requireVerification: getEnvBoolWithDefault(
			"AUTH_REQUIRE_EMAIL_VERIFICATION",
			IsProduction(),
		),
		verificationTTL: getEnvDurationWithDefault(
			"AUTH_EMAIL_VERIFICATION_TTL",
			48*time.Hour,
		),
		verificationURL: getEnvWithDefault("AUTH_EMAIL_VERIFICATION_URL", ""),
		resetTTL:        getEnvDurationWithDefault("AUTH_PASSWORD_RESET_TTL", time.Hour),
		resetURL:        getEnvWithDefault("AUTH_PASSWORD_RESET_URL", ""),
		senderAddress:   getEnvWithDefault("SYSTEM_MAIL_ADDRESS", ""),
		senderName:      getEnvWithDefault("SYSTEM_MAIL_NAME", "AutoReach"),
		smtpHost:        getEnvWithDefault("SYSTEM_SMTP_HOST", ""),
		smtpPort:        getEnvIntWithDefault("SYSTEM_SMTP_PORT", 587),
		smtpSecurity:    getEnvWithDefault("SYSTEM_SMTP_SECURITY", "STARTTLS"),
		smtpUsername:    getEnvWithDefault("SYSTEM_SMTP_USERNAME", ""),
		smtpPassword:    getEnvWithDefault("SYSTEM_SMTP_PASSWORD", ""),
	}

	if accountEmailConfigI.requireVerification &&
		(!IsSystemMailConfigured() || accountEmailConfigI.resetURL == "" ||
			userConfigI.emailChangeURL == "") {
		log.Fatalf("AUTH_REQUIRE_EMAIL_VERIFICATION needs AUTH_EMAIL_VERIFICATION_URL, " +
			"AUTH_PASSWORD_RESET_URL, AUTH_EMAIL_CHANGE_URL, SYSTEM_MAIL_ADDRESS and SYSTEM_SMTP_HOST")
	}
}

func IsEmailVerificationRequired() bool { return accountEmailConfigI.requireVerification }

func IsSystemMailConfigured() bool {
	return accountEmailConfigI.verificationURL != "" && accountEmailConfigI.senderAddress != "" &&
		accountEmailConfigI.smtpHost != ""
}

func GetEmailVerificationTTL() time.Duration { return accountEmailConfigI.verificationTTL }
func GetEmailVerificationURL() string        { return accountEmailConfigI.verificationURL }
func GetEmailVerificationKeyLabel() string   { return emailVerificationKeyLabel }
func GetPasswordResetTTL() time.Duration     { return accountEmailConfigI.resetTTL }
func GetPasswordResetURL() string            { return accountEmailConfigI.resetURL }
func GetPasswordResetKeyLabel() string       { return passwordResetKeyLabel }
func GetAccountTokenSeparator() string       { return accountTokenSeparator }
func GetAccountTokenParam() string           { return accountTokenParam }
func GetAccountTokenParts() int              { return accountTokenParts }
func GetSystemMailAddress() string           { return accountEmailConfigI.senderAddress }
func GetSystemMailName() string              { return accountEmailConfigI.senderName }
func GetSystemSMTPHost() string              { return accountEmailConfigI.smtpHost }
func GetSystemSMTPPort() int                 { return accountEmailConfigI.smtpPort }
func GetSystemSMTPSecurity() string          { return accountEmailConfigI.smtpSecurity }
func GetSystemSMTPUsername() string          { return accountEmailConfigI.smtpUsername }
func GetSystemSMTPPassword() string          { return accountEmailConfigI.smtpPassword }
