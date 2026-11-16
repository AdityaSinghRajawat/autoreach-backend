package config

import (
	"log"
	"time"
)

const (
	emailVerificationKeyLabel  = "email-verification"
	verificationTokenSeparator = "."
	verificationTokenParam     = "token"
	verificationTokenParts     = 3
)

type signupConfig struct {
	requireVerification bool
	verificationTTL     time.Duration
	verificationURL     string
	senderAddress       string
	senderName          string
	smtpHost            string
	smtpPort            int
	smtpSecurity        string
	smtpUsername        string
	smtpPassword        string
}

var signupConfigI signupConfig

func loadSignupConfig() {
	signupConfigI = signupConfig{
		requireVerification: getEnvBoolWithDefault(
			"AUTH_REQUIRE_EMAIL_VERIFICATION",
			IsProduction(),
		),
		verificationTTL: getEnvDurationWithDefault(
			"AUTH_EMAIL_VERIFICATION_TTL",
			48*time.Hour,
		),
		verificationURL: getEnvWithDefault("AUTH_EMAIL_VERIFICATION_URL", ""),
		senderAddress:   getEnvWithDefault("SYSTEM_MAIL_ADDRESS", ""),
		senderName:      getEnvWithDefault("SYSTEM_MAIL_NAME", "AutoReach"),
		smtpHost:        getEnvWithDefault("SYSTEM_SMTP_HOST", ""),
		smtpPort:        getEnvIntWithDefault("SYSTEM_SMTP_PORT", 587),
		smtpSecurity:    getEnvWithDefault("SYSTEM_SMTP_SECURITY", "STARTTLS"),
		smtpUsername:    getEnvWithDefault("SYSTEM_SMTP_USERNAME", ""),
		smtpPassword:    getEnvWithDefault("SYSTEM_SMTP_PASSWORD", ""),
	}

	if signupConfigI.requireVerification && !IsSystemMailConfigured() {
		log.Fatalf("AUTH_REQUIRE_EMAIL_VERIFICATION needs AUTH_EMAIL_VERIFICATION_URL, " +
			"SYSTEM_MAIL_ADDRESS and SYSTEM_SMTP_HOST")
	}
}

// IsEmailVerificationRequired hides whether an address is registered: signup answers the same for
// new and existing addresses and the account works only after the emailed link is opened.
func IsEmailVerificationRequired() bool { return signupConfigI.requireVerification }

func IsSystemMailConfigured() bool {
	return signupConfigI.verificationURL != "" && signupConfigI.senderAddress != "" &&
		signupConfigI.smtpHost != ""
}

func GetEmailVerificationTTL() time.Duration { return signupConfigI.verificationTTL }
func GetEmailVerificationURL() string        { return signupConfigI.verificationURL }
func GetEmailVerificationKeyLabel() string   { return emailVerificationKeyLabel }
func GetSystemMailAddress() string           { return signupConfigI.senderAddress }
func GetSystemMailName() string              { return signupConfigI.senderName }
func GetSystemSMTPHost() string              { return signupConfigI.smtpHost }
func GetSystemSMTPPort() int                 { return signupConfigI.smtpPort }
func GetSystemSMTPSecurity() string          { return signupConfigI.smtpSecurity }
func GetSystemSMTPUsername() string          { return signupConfigI.smtpUsername }
func GetSystemSMTPPassword() string          { return signupConfigI.smtpPassword }
func GetVerificationTokenSeparator() string  { return verificationTokenSeparator }
func GetVerificationTokenParam() string      { return verificationTokenParam }
func GetVerificationTokenParts() int         { return verificationTokenParts }
