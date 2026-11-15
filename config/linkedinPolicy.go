package config

import "time"

const (
	linkedinNoteMaxLength    = 300
	linkedinProviderAssisted = "assisted"
	linkedinInviteLockPrefix = "linkedin_invites:"
)

type linkedinConfig struct {
	provider          string
	dailyInviteLimit  int
	acceptanceWait    time.Duration
	acceptanceRecheck time.Duration
	capRetryDelay     time.Duration
}

var linkedinConfigI linkedinConfig

func loadLinkedinConfig() {
	linkedinConfigI = linkedinConfig{
		provider:          getEnvWithDefault("LINKEDIN_PROVIDER", linkedinProviderAssisted),
		dailyInviteLimit:  max(getEnvIntWithDefault("LINKEDIN_DAILY_INVITE_LIMIT", 20), 1),
		acceptanceWait:    getEnvDurationWithDefault("LINKEDIN_ACCEPTANCE_WAIT", 30*24*time.Hour),
		acceptanceRecheck: getEnvDurationWithDefault("LINKEDIN_ACCEPTANCE_RECHECK", 24*time.Hour),
		capRetryDelay:     getEnvDurationWithDefault("LINKEDIN_CAP_RETRY_DELAY", time.Hour),
	}
}

func GetLinkedinNoteMaxLength() int               { return linkedinNoteMaxLength }
func GetLinkedinProviderAssisted() string         { return linkedinProviderAssisted }
func GetLinkedinInviteLockPrefix() string         { return linkedinInviteLockPrefix }
func GetLinkedinProvider() string                 { return linkedinConfigI.provider }
func GetLinkedinDailyInviteLimit() int            { return linkedinConfigI.dailyInviteLimit }
func GetLinkedinAcceptanceWait() time.Duration    { return linkedinConfigI.acceptanceWait }
func GetLinkedinAcceptanceRecheck() time.Duration { return linkedinConfigI.acceptanceRecheck }
func GetLinkedinCapRetryDelay() time.Duration     { return linkedinConfigI.capRetryDelay }
