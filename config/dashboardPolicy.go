package config

import "time"

const (
	activityCursorSeparator = "|"
	activityDefaultLimit    = 20
	activityStreamBatchSize = 50
	sseContentType          = "text/event-stream"
	sseActivityEvent        = "activity"
	lastEventIDHeader       = "Last-Event-ID"
	proxyBufferingHeader    = "X-Accel-Buffering"
	proxyBufferingOff       = "no"
	candidateIDQueryParam   = "candidateId"
	queueParam              = "queue"
	taskIDParam             = "taskId"
	suppressionIDParam      = "suppressionId"
	entryTypeFieldName      = "EntryType"
	deliverabilityWindow    = 7 * 24 * time.Hour
	percentageScale         = 100
)

type dashboardConfig struct {
	streamPollInterval time.Duration
	streamHeartbeat    time.Duration
	streamMaxDuration  time.Duration
	streamsPerOwner    int
}

var dashboardConfigI dashboardConfig

func loadDashboardConfig() {
	dashboardConfigI = dashboardConfig{
		streamPollInterval: getEnvDurationWithDefault(
			"ACTIVITY_STREAM_POLL_INTERVAL",
			5*time.Second,
		),
		streamHeartbeat: getEnvDurationWithDefault("ACTIVITY_STREAM_HEARTBEAT", 25*time.Second),
		streamMaxDuration: getEnvDurationWithDefault(
			"ACTIVITY_STREAM_MAX_DURATION",
			30*time.Minute,
		),
		streamsPerOwner: max(getEnvIntWithDefault("ACTIVITY_STREAMS_PER_CANDIDATE", 5), 1),
	}
}

func GetActivityCursorSeparator() string     { return activityCursorSeparator }
func GetActivityDefaultLimit() int           { return activityDefaultLimit }
func GetActivityStreamBatchSize() int        { return activityStreamBatchSize }
func GetSSEContentType() string              { return sseContentType }
func GetSSEActivityEvent() string            { return sseActivityEvent }
func GetLastEventIDHeader() string           { return lastEventIDHeader }
func GetProxyBufferingHeader() string        { return proxyBufferingHeader }
func GetProxyBufferingOff() string           { return proxyBufferingOff }
func GetCandidateIDQueryParam() string       { return candidateIDQueryParam }
func GetQueueParam() string                  { return queueParam }
func GetTaskIDParam() string                 { return taskIDParam }
func GetSuppressionIDParam() string          { return suppressionIDParam }
func GetEntryTypeFieldName() string          { return entryTypeFieldName }
func GetDeliverabilityWindow() time.Duration { return deliverabilityWindow }
func GetPercentageScale() float64            { return percentageScale }

func GetActivityStreamPollInterval() time.Duration { return dashboardConfigI.streamPollInterval }
func GetActivityStreamHeartbeat() time.Duration    { return dashboardConfigI.streamHeartbeat }
func GetActivityStreamMaxDuration() time.Duration  { return dashboardConfigI.streamMaxDuration }

func GetActivityStreamsPerCandidate() int { return dashboardConfigI.streamsPerOwner }
