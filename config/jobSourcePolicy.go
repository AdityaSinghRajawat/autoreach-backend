package config

import "time"

const (
	jobSourceIDParam         = "sourceId"
	jobSourcePollerLock      = "job_source_poller"
	jobSourceRegionEU        = "EU"
	greenhouseBoardJobsQuery = "?content=true"
	leverPostingsQuery       = "?mode=json"
	maxJobSourceError        = 500
	ingestBatchSize          = 500
)

type jobSourceConfig struct {
	pollInterval time.Duration
}

var jobSourceConfigI jobSourceConfig

func loadJobSourceConfig() {
	jobSourceConfigI = jobSourceConfig{
		pollInterval: getEnvDurationWithDefault("JOB_SOURCE_POLL_INTERVAL", time.Hour),
	}
}

func GetJobSourceIDParam() string             { return jobSourceIDParam }
func GetJobSourcePollerLock() string          { return jobSourcePollerLock }
func GetJobSourceRegionEU() string            { return jobSourceRegionEU }
func GetGreenhouseBoardJobsQuery() string     { return greenhouseBoardJobsQuery }
func GetLeverPostingsQuery() string           { return leverPostingsQuery }
func GetMaxJobSourceErrorLength() int         { return maxJobSourceError }
func GetIngestBatchSize() int                 { return ingestBatchSize }
func GetJobSourcePollInterval() time.Duration { return jobSourceConfigI.pollInterval }
