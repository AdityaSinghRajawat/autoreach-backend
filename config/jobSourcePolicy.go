package config

import (
	"strings"
	"time"
)

const (
	jobSourceIDParam            = "sourceId"
	jobSourcePollerLock         = "job_source_poller"
	jobSourceRegionEU           = "EU"
	greenhouseBoardJobsQuery    = "?content=true"
	greenhouseJobURLFormat      = "https://job-boards.greenhouse.io/%s/jobs/%d"
	leverPostingsQuery          = "?mode=json"
	ashbyJobBoardPath           = "/job-board/"
	smartRecruitersPageSize     = 100
	smartRecruitersPostingsPath = "/companies/%s/postings"
	maxJobSourceError           = 500
	ingestBatchSize             = 500
)

type jobSourceConfig struct {
	pollInterval              time.Duration
	boardTimeout              time.Duration
	ashbyAPIURL               string
	smartRecruitersAPIURL     string
	smartRecruitersMaxDetails int
}

var jobSourceConfigI jobSourceConfig

func loadJobSourceConfig() {
	jobSourceConfigI = jobSourceConfig{
		pollInterval: getEnvDurationWithDefault("JOB_SOURCE_POLL_INTERVAL", time.Hour),
		boardTimeout: getEnvDurationWithDefault("JOB_BOARD_HTTP_TIMEOUT", time.Minute),
		ashbyAPIURL: strings.TrimRight(
			getEnvWithDefault("ASHBY_JOB_BOARD_API_URL", "https://api.ashbyhq.com/posting-api"),
			"/",
		),
		smartRecruitersAPIURL: strings.TrimRight(
			getEnvWithDefault("SMARTRECRUITERS_API_URL", "https://api.smartrecruiters.com/v1"),
			"/",
		),
		smartRecruitersMaxDetails: getEnvIntWithDefault(
			"SMARTRECRUITERS_MAX_DETAILS_PER_POLL",
			100,
		),
	}
}

func GetJobSourceIDParam() string              { return jobSourceIDParam }
func GetJobSourcePollerLock() string           { return jobSourcePollerLock }
func GetJobSourceRegionEU() string             { return jobSourceRegionEU }
func GetGreenhouseBoardJobsQuery() string      { return greenhouseBoardJobsQuery }
func GetGreenhouseJobURLFormat() string        { return greenhouseJobURLFormat }
func GetLeverPostingsQuery() string            { return leverPostingsQuery }
func GetAshbyJobBoardPath() string             { return ashbyJobBoardPath }
func GetSmartRecruitersPageSize() int          { return smartRecruitersPageSize }
func GetSmartRecruitersPostingsPath() string   { return smartRecruitersPostingsPath }
func GetMaxJobSourceErrorLength() int          { return maxJobSourceError }
func GetIngestBatchSize() int                  { return ingestBatchSize }
func GetJobSourcePollInterval() time.Duration  { return jobSourceConfigI.pollInterval }
func GetJobBoardTimeout() time.Duration        { return jobSourceConfigI.boardTimeout }
func GetAshbyJobBoardAPIURL() string           { return jobSourceConfigI.ashbyAPIURL }
func GetSmartRecruitersAPIURL() string         { return jobSourceConfigI.smartRecruitersAPIURL }
func GetSmartRecruitersMaxDetailsPerPoll() int { return jobSourceConfigI.smartRecruitersMaxDetails }
