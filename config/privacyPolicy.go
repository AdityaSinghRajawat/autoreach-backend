package config

import (
	"log"
	"time"
)

const (
	dataRetentionLock     = "data_retention"
	contactIDParam        = "contactId"
	contactSignalIDParam  = "signalId"
	contactHistoryKey     = "history"
	contactHistoryLimit   = 20
	anonymizedContactName = "Former contact"
)

type privacyConfig struct {
	contactRetention    time.Duration
	retentionCheck      time.Duration
	tokenRetention      time.Duration
	eventRetention      time.Duration
	signalConfirmations int
}

var privacyConfigI privacyConfig

func loadPrivacyConfig() {
	confirmations := getEnvIntWithDefault("CONTACT_SIGNAL_CONFIRMATIONS", 2)
	if confirmations < 1 {
		log.Fatalf("CONTACT_SIGNAL_CONFIRMATIONS must be 1 or more")
	}

	privacyConfigI = privacyConfig{
		signalConfirmations: confirmations,
		contactRetention:    getEnvDurationWithDefault("CONTACT_RETENTION_PERIOD", 0),
		retentionCheck: getEnvDurationWithDefault(
			"DATA_RETENTION_CHECK_INTERVAL",
			getEnvDurationWithDefault("CONTACT_RETENTION_CHECK_INTERVAL", 24*time.Hour),
		),
		tokenRetention: getEnvDurationWithDefault("REFRESH_TOKEN_RETENTION", 7*24*time.Hour),
		eventRetention: getEnvDurationWithDefault("OUTREACH_EVENT_RETENTION", 0),
	}
}

func GetDataRetentionLock() string     { return dataRetentionLock }
func GetContactIDParam() string        { return contactIDParam }
func GetContactSignalIDParam() string  { return contactSignalIDParam }
func GetContactHistoryKey() string     { return contactHistoryKey }
func GetContactHistoryLimit() int      { return contactHistoryLimit }
func GetAnonymizedContactName() string { return anonymizedContactName }

// GetContactRetentionPeriod is how long an unused contact keeps its personal details; zero keeps
// them forever.
func GetContactRetentionPeriod() time.Duration { return privacyConfigI.contactRetention }

func GetDataRetentionCheckInterval() time.Duration { return privacyConfigI.retentionCheck }

// GetRefreshTokenRetention is how long an expired refresh token is kept; reuse of a token this
// old is refused as unknown instead of revoking its family, which no longer matters once it expired.
func GetRefreshTokenRetention() time.Duration { return privacyConfigI.tokenRetention }

// GetOutreachEventRetention is how long opens, clicks and other outreach events are kept; zero keeps
// them forever. Removing events also removes them from dashboard history.
func GetOutreachEventRetention() time.Duration { return privacyConfigI.eventRetention }

// GetContactSignalConfirmations is how many candidates must observe the same bounce, reply or
// LinkedIn response before it changes the contact they share.
func GetContactSignalConfirmations() int { return privacyConfigI.signalConfirmations }
