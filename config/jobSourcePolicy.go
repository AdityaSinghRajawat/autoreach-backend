package config

import (
	"regexp"
	"strings"
	"time"
)

const (
	jobSourceIDParam            = "sourceId"
	jobSourcePollerLock         = "job_source_poller"
	jobSourceRegionEU           = "EU"
	boardTypeFieldName          = "BoardType"
	careerPageBoardType         = "CAREER_PAGE"
	otherATSType                = "OTHER"
	greenhouseBoardJobsQuery    = "?content=true"
	greenhouseJobURLFormat      = "https://job-boards.greenhouse.io/%s/jobs/%d"
	leverPostingsQuery          = "?mode=json"
	ashbyJobBoardPath           = "/job-board/"
	smartRecruitersPageSize     = 100
	smartRecruitersPostingsPath = "/companies/%s/postings"
	workdayPageSize             = 20
	workdayMaxListed            = 2000
	workdayJobsPathFormat       = "/wday/cxs/%s/%s/jobs"
	workdayJobPathFormat        = "/wday/cxs/%s/%s%s"
	workdayTokenSeparator       = "/"
	workdayDateLayout           = "2006-01-02"
	adzunaSearchPathFormat      = "/jobs/%s/search/%d"
	adzunaPredictedSalary       = "1"
	careerPageMaxSitemaps       = 10
	careerPageMaxBytes          = 10 << 20
	careerPageRobotsPath        = "/robots.txt"
	userAgentHeader             = "User-Agent"
	maxExternalJobIDLength      = 255
	maxJobSourceError           = 500
	ingestBatchSize             = 500
	companyDomainLookupCredits  = 1
	companyDomainCandidates     = 5
	hunterMinCompanyNameLength  = 3
)

type boardJobKinds struct {
	jobSource string
	atsType   string
}

var boardJobKindsI = map[string]boardJobKinds{
	"GREENHOUSE":        {jobSource: "GREENHOUSE", atsType: "GREENHOUSE"},
	"LEVER":             {jobSource: "LEVER", atsType: "LEVER"},
	"ASHBY":             {jobSource: "ASHBY", atsType: "ASHBY"},
	"SMARTRECRUITERS":   {jobSource: "SMARTRECRUITERS", atsType: "SMARTRECRUITERS"},
	"WORKDAY":           {jobSource: "WORKDAY", atsType: "WORKDAY"},
	"ADZUNA":            {jobSource: "ADZUNA", atsType: otherATSType},
	careerPageBoardType: {jobSource: "DIRECT", atsType: otherATSType},
}

var adzunaCurrenciesI = map[string]string{
	"at": "EUR", "au": "AUD", "be": "EUR", "br": "BRL", "ca": "CAD", "ch": "CHF", "de": "EUR",
	"es": "EUR", "fr": "EUR", "gb": "GBP", "in": "INR", "it": "EUR", "mx": "MXN", "nl": "EUR",
	"nz": "NZD", "pl": "PLN", "sg": "SGD", "us": "USD", "za": "ZAR",
}

var boardTokenPatternsI = map[string]*regexp.Regexp{
	"GREENHOUSE":      regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`),
	"LEVER":           regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`),
	"ASHBY":           regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`),
	"SMARTRECRUITERS": regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`),
	"WORKDAY":         regexp.MustCompile(`^[a-z0-9-]+\.wd[0-9]+/[a-z0-9_-]+$`),
	"ADZUNA": regexp.MustCompile(
		`^(at|au|be|br|ca|ch|de|es|fr|gb|in|it|mx|nl|nz|pl|sg|us|za)$`,
	),
}

type jobSourceConfig struct {
	pollInterval            time.Duration
	staleAfter              time.Duration
	boardTimeout            time.Duration
	maxDetailsPerPoll       int
	detailConcurrency       int
	ashbyAPIURL             string
	smartRecruitersAPIURL   string
	workdayHostURLFormat    string
	adzunaAPIURL            string
	adzunaAppID             string
	adzunaAppKey            string
	adzunaResultsPerPage    int
	adzunaMaxPages          int
	careerPageUserAgent     string
	companyDomainLookups    int
	companyDomainRetryAfter time.Duration
}

var jobSourceConfigI jobSourceConfig

func loadJobSourceConfig() {
	jobSourceConfigI = jobSourceConfig{
		pollInterval:      getEnvDurationWithDefault("JOB_SOURCE_POLL_INTERVAL", time.Hour),
		staleAfter:        getEnvDurationWithDefault("JOB_SOURCE_STALE_AFTER", 30*24*time.Hour),
		boardTimeout:      getEnvDurationWithDefault("JOB_BOARD_HTTP_TIMEOUT", time.Minute),
		maxDetailsPerPoll: getEnvIntWithDefault("JOB_BOARD_MAX_DETAILS_PER_POLL", 100),
		detailConcurrency: max(getEnvIntWithDefault("JOB_BOARD_DETAIL_CONCURRENCY", 4), 1),
		ashbyAPIURL: strings.TrimRight(
			getEnvWithDefault("ASHBY_JOB_BOARD_API_URL", "https://api.ashbyhq.com/posting-api"),
			"/",
		),
		smartRecruitersAPIURL: strings.TrimRight(
			getEnvWithDefault("SMARTRECRUITERS_API_URL", "https://api.smartrecruiters.com/v1"),
			"/",
		),
		workdayHostURLFormat: getEnvWithDefault(
			"WORKDAY_HOST_URL_FORMAT",
			"https://%s.%s.myworkdayjobs.com",
		),
		adzunaAPIURL: strings.TrimRight(
			getEnvWithDefault("ADZUNA_API_URL", "https://api.adzuna.com/v1/api"),
			"/",
		),
		adzunaAppID:          getEnvWithDefault("ADZUNA_APP_ID", ""),
		adzunaAppKey:         getEnvWithDefault("ADZUNA_APP_KEY", ""),
		adzunaResultsPerPage: getEnvIntWithDefault("ADZUNA_RESULTS_PER_PAGE", 50),
		adzunaMaxPages:       getEnvIntWithDefault("ADZUNA_MAX_PAGES", 4),
		careerPageUserAgent:  getEnvWithDefault("CAREER_PAGE_USER_AGENT", "AutoReachBot/1.0"),
		companyDomainLookups: getEnvIntWithDefault("COMPANY_DOMAIN_LOOKUPS_PER_POLL", 50),
		companyDomainRetryAfter: getEnvDurationWithDefault(
			"COMPANY_DOMAIN_RETRY_AFTER",
			7*24*time.Hour,
		),
	}
}

func GetJobSourceIDParam() string            { return jobSourceIDParam }
func GetJobSourcePollerLock() string         { return jobSourcePollerLock }
func GetJobSourceRegionEU() string           { return jobSourceRegionEU }
func GetBoardTypeFieldName() string          { return boardTypeFieldName }
func GetCareerPageBoardType() string         { return careerPageBoardType }
func GetOtherATSType() string                { return otherATSType }
func GetGreenhouseBoardJobsQuery() string    { return greenhouseBoardJobsQuery }
func GetGreenhouseJobURLFormat() string      { return greenhouseJobURLFormat }
func GetLeverPostingsQuery() string          { return leverPostingsQuery }
func GetAshbyJobBoardPath() string           { return ashbyJobBoardPath }
func GetSmartRecruitersPageSize() int        { return smartRecruitersPageSize }
func GetSmartRecruitersPostingsPath() string { return smartRecruitersPostingsPath }
func GetWorkdayPageSize() int                { return workdayPageSize }
func GetWorkdayMaxListed() int               { return workdayMaxListed }
func GetWorkdayJobsPathFormat() string       { return workdayJobsPathFormat }
func GetWorkdayJobPathFormat() string        { return workdayJobPathFormat }
func GetWorkdayTokenSeparator() string       { return workdayTokenSeparator }
func GetWorkdayDateLayout() string           { return workdayDateLayout }
func GetAdzunaSearchPathFormat() string      { return adzunaSearchPathFormat }
func GetAdzunaPredictedSalary() string       { return adzunaPredictedSalary }
func GetCareerPageMaxSitemaps() int          { return careerPageMaxSitemaps }
func GetCareerPageMaxBytes() int64           { return careerPageMaxBytes }
func GetCareerPageRobotsPath() string        { return careerPageRobotsPath }
func GetUserAgentHeader() string             { return userAgentHeader }
func GetMaxExternalJobIDLength() int         { return maxExternalJobIDLength }
func GetMaxJobSourceErrorLength() int        { return maxJobSourceError }
func GetIngestBatchSize() int                { return ingestBatchSize }
func GetCompanyDomainCandidates() int        { return companyDomainCandidates }
func GetHunterMinCompanyNameLength() int     { return hunterMinCompanyNameLength }

func GetJobSourcePollInterval() time.Duration   { return jobSourceConfigI.pollInterval }
func GetJobSourceStaleAfter() time.Duration     { return jobSourceConfigI.staleAfter }
func GetJobBoardTimeout() time.Duration         { return jobSourceConfigI.boardTimeout }
func GetJobBoardMaxDetailsPerPoll() int         { return jobSourceConfigI.maxDetailsPerPoll }
func GetAshbyJobBoardAPIURL() string            { return jobSourceConfigI.ashbyAPIURL }
func GetSmartRecruitersAPIURL() string          { return jobSourceConfigI.smartRecruitersAPIURL }
func GetWorkdayHostURLFormat() string           { return jobSourceConfigI.workdayHostURLFormat }
func GetAdzunaAPIURL() string                   { return jobSourceConfigI.adzunaAPIURL }
func GetAdzunaAppID() string                    { return jobSourceConfigI.adzunaAppID }
func GetAdzunaAppKey() string                   { return jobSourceConfigI.adzunaAppKey }
func GetAdzunaResultsPerPage() int              { return jobSourceConfigI.adzunaResultsPerPage }
func GetAdzunaMaxPages() int                    { return jobSourceConfigI.adzunaMaxPages }
func GetCareerPageUserAgent() string            { return jobSourceConfigI.careerPageUserAgent }
func GetCompanyDomainLookupsPerPoll() int       { return jobSourceConfigI.companyDomainLookups }
func GetCompanyDomainRetryAfter() time.Duration { return jobSourceConfigI.companyDomainRetryAfter }

// GetBoardJobKinds returns the job source and ATS type recorded on jobs read from a board type.
func GetBoardJobKinds(boardType string) (string, string) {
	kinds := boardJobKindsI[boardType]

	return kinds.jobSource, kinds.atsType
}

func GetBoardTokenPattern(boardType string) (*regexp.Regexp, bool) {
	pattern, found := boardTokenPatternsI[boardType]

	return pattern, found
}

func GetAdzunaCurrency(country string) (string, bool) {
	currency, found := adzunaCurrenciesI[country]

	return currency, found
}

// Kept low: every read goes to the same employer's site.
func GetJobBoardDetailConcurrency() int { return jobSourceConfigI.detailConcurrency }
