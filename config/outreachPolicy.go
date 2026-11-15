package config

import "time"

const (
	messageIDParam           = "outreachId"
	outreachDispatchLock     = "outreach_dispatcher"
	replySweeperLock         = "reply_sweeper"
	sendAttemptWindow        = 24 * time.Hour
	skipAutoOutreachDisabled = "AUTO_OUTREACH_DISABLED"
	skipNoActiveCampaign     = "NO_ACTIVE_CAMPAIGN"
	skipNoActiveMailbox      = "NO_ACTIVE_MAILBOX"
	skipTemplateNotApproved  = "TEMPLATE_NOT_APPROVED"
	bounceReason             = "the recipient's mail server rejected the message"
)

type outreachConfig struct {
	sendWindowStartHour int
	sendWindowEndHour   int
	sendOnWeekends      bool
	sendJitter          time.Duration
	contactsPerJob      int
	dispatchBatchSize   int
	dispatchInterval    time.Duration
	replySweepInterval  time.Duration
	maxSendAttempts     int
	sendRetryDelay      time.Duration
	sendingStaleAfter   time.Duration
}

var outreachConfigI outreachConfig

func loadOutreachConfig() {
	outreachConfigI = outreachConfig{
		sendWindowStartHour: min(
			max(getEnvIntWithDefault("OUTREACH_SEND_WINDOW_START_HOUR", 9), 0),
			23,
		),
		sendWindowEndHour: min(
			max(getEnvIntWithDefault("OUTREACH_SEND_WINDOW_END_HOUR", 17), 1),
			24,
		),
		sendOnWeekends:    getEnvBoolWithDefault("OUTREACH_SEND_ON_WEEKENDS", false),
		sendJitter:        getEnvDurationWithDefault("OUTREACH_SEND_JITTER", 10*time.Minute),
		contactsPerJob:    max(getEnvIntWithDefault("OUTREACH_CONTACTS_PER_JOB", 1), 1),
		dispatchBatchSize: max(getEnvIntWithDefault("OUTREACH_DISPATCH_BATCH_SIZE", 100), 1),
		dispatchInterval:  getEnvDurationWithDefault("OUTREACH_DISPATCH_INTERVAL", time.Minute),
		replySweepInterval: getEnvDurationWithDefault(
			"OUTREACH_REPLY_SWEEP_INTERVAL",
			15*time.Minute,
		),
		maxSendAttempts: max(getEnvIntWithDefault("OUTREACH_MAX_SEND_ATTEMPTS", 3), 1),
		sendRetryDelay:  getEnvDurationWithDefault("OUTREACH_SEND_RETRY_DELAY", 15*time.Minute),
		sendingStaleAfter: getEnvDurationWithDefault(
			"OUTREACH_SENDING_STALE_AFTER",
			15*time.Minute,
		),
	}
}

func GetMessageIDParam() string                  { return messageIDParam }
func GetOutreachDispatchLock() string            { return outreachDispatchLock }
func GetReplySweeperLock() string                { return replySweeperLock }
func GetSendAttemptWindow() time.Duration        { return sendAttemptWindow }
func GetSkipAutoOutreachDisabled() string        { return skipAutoOutreachDisabled }
func GetSkipNoActiveCampaign() string            { return skipNoActiveCampaign }
func GetSkipNoActiveMailbox() string             { return skipNoActiveMailbox }
func GetSkipTemplateNotApproved() string         { return skipTemplateNotApproved }
func GetBounceReason() string                    { return bounceReason }
func GetSendWindowStartHour() int                { return outreachConfigI.sendWindowStartHour }
func GetSendWindowEndHour() int                  { return outreachConfigI.sendWindowEndHour }
func IsSendingOnWeekends() bool                  { return outreachConfigI.sendOnWeekends }
func GetSendJitter() time.Duration               { return outreachConfigI.sendJitter }
func GetOutreachContactsPerJob() int             { return outreachConfigI.contactsPerJob }
func GetOutreachDispatchBatchSize() int          { return outreachConfigI.dispatchBatchSize }
func GetOutreachDispatchInterval() time.Duration { return outreachConfigI.dispatchInterval }
func GetReplySweepInterval() time.Duration       { return outreachConfigI.replySweepInterval }
func GetMaxSendAttempts() int                    { return outreachConfigI.maxSendAttempts }
func GetSendRetryDelay() time.Duration           { return outreachConfigI.sendRetryDelay }
func GetSendingStaleAfter() time.Duration        { return outreachConfigI.sendingStaleAfter }
