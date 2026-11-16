package config

import (
	"log"
	"time"
)

const (
	contactRetentionLock  = "contact_retention"
	contactIDParam        = "contactId"
	contactSignalIDParam  = "signalId"
	contactHistoryKey     = "history"
	contactHistoryLimit   = 20
	anonymizedContactName = "Former contact"
)

type privacyConfig struct {
	contactRetention      time.Duration
	contactRetentionCheck time.Duration
	signalConfirmations   int
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
		contactRetentionCheck: getEnvDurationWithDefault(
			"CONTACT_RETENTION_CHECK_INTERVAL",
			24*time.Hour,
		),
	}
}

func GetContactRetentionLock() string  { return contactRetentionLock }
func GetContactIDParam() string        { return contactIDParam }
func GetContactSignalIDParam() string  { return contactSignalIDParam }
func GetContactHistoryKey() string     { return contactHistoryKey }
func GetContactHistoryLimit() int      { return contactHistoryLimit }
func GetAnonymizedContactName() string { return anonymizedContactName }

// GetContactRetentionPeriod is how long an unused contact keeps its personal details; zero keeps
// them forever.
func GetContactRetentionPeriod() time.Duration { return privacyConfigI.contactRetention }

func GetContactRetentionCheckInterval() time.Duration { return privacyConfigI.contactRetentionCheck }

// GetContactSignalConfirmations is how many candidates must observe the same bounce, reply or
// LinkedIn response before it changes the contact they share.
func GetContactSignalConfirmations() int { return privacyConfigI.signalConfirmations }
