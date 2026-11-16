package config

import "time"

const (
	contactRetentionLock  = "contact_retention"
	contactIDParam        = "contactId"
	contactHistoryKey     = "history"
	contactHistoryLimit   = 20
	anonymizedContactName = "Former contact"
)

type privacyConfig struct {
	contactRetention      time.Duration
	contactRetentionCheck time.Duration
}

var privacyConfigI privacyConfig

func loadPrivacyConfig() {
	privacyConfigI = privacyConfig{
		contactRetention: getEnvDurationWithDefault("CONTACT_RETENTION_PERIOD", 0),
		contactRetentionCheck: getEnvDurationWithDefault(
			"CONTACT_RETENTION_CHECK_INTERVAL",
			24*time.Hour,
		),
	}
}

func GetContactRetentionLock() string  { return contactRetentionLock }
func GetContactIDParam() string        { return contactIDParam }
func GetContactHistoryKey() string     { return contactHistoryKey }
func GetContactHistoryLimit() int      { return contactHistoryLimit }
func GetAnonymizedContactName() string { return anonymizedContactName }

// GetContactRetentionPeriod is how long an unused contact keeps its personal details; zero keeps
// them forever.
func GetContactRetentionPeriod() time.Duration { return privacyConfigI.contactRetention }

func GetContactRetentionCheckInterval() time.Duration { return privacyConfigI.contactRetentionCheck }
