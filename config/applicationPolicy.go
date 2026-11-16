package config

import (
	"log"
	"strings"
	"time"
)

const (
	applicationIDParam         = "applicationId"
	applicationStepSubmitting  = "SUBMITTING"
	applicationStepSubmitted   = "SUBMITTED"
	applicationProfileKey      = "applicationProfile"
	atsRateLimitName           = "ats"
	greenhouseQuestionsQuery   = "?questions=true"
	leverSiteKeyQuery          = "?key="
	answerYes                  = "Yes"
	answerNo                   = "No"
	formHasHiddenQuestionsNote = "This posting's form may have extra questions that are only visible on the form itself."
	unsupportedATSNote         = "AutoReach cannot read this application form; apply with the link."
	keyPairSeparator           = ":"
	greenhouseHostSuffix       = "greenhouse.io"
	leverHostSuffix            = "lever.co"
	leverEUHostMarker          = ".eu."
	greenhouseJobsSegment      = "jobs"
	leverApplySegment          = "apply"
	leverLinkedinURLField      = "urls[LinkedIn]"
	leverGithubURLField        = "urls[GitHub]"
	leverPortfolioURLField     = "urls[Portfolio]"
)

// Each profile answer is matched to a form question by keywords in the question's label.
const (
	ProfileFieldLinkedin        = "linkedinUrl"
	ProfileFieldGithub          = "githubUrl"
	ProfileFieldWebsite         = "websiteUrl"
	ProfileFieldWorkAuth        = "workAuthorized"
	ProfileFieldSponsorship     = "requiresSponsorship"
	ProfileFieldSalary          = "salaryExpectation"
	ProfileFieldNoticePeriod    = "noticePeriod"
	ProfileFieldHeardFrom       = "howDidYouHear"
	ProfileFieldCurrentEmployer = "currentCompany"
	ProfileFieldLocation        = "location"
)

var profileFieldKeywordsI = []struct {
	Field    string
	Keywords []string
}{
	{ProfileFieldLinkedin, []string{"linkedin"}},
	{ProfileFieldGithub, []string{"github"}},
	{ProfileFieldWebsite, []string{"website", "portfolio", "personal site"}},
	{ProfileFieldSponsorship, []string{"sponsor"}},
	{
		ProfileFieldWorkAuth,
		[]string{
			"authorized to work",
			"authorised to work",
			"eligible to work",
			"legally authorized",
			"work authorization",
			"right to work",
		},
	},
	{ProfileFieldSalary, []string{"salary", "compensation"}},
	{ProfileFieldNoticePeriod, []string{"notice period", "start date", "when can you start"}},
	{ProfileFieldHeardFrom, []string{"hear about"}},
	{ProfileFieldCurrentEmployer, []string{"current company", "current employer"}},
	{ProfileFieldLocation, []string{"location", "where are you based", "city"}},
}

type applicationConfig struct {
	greenhouseAPIURL string
	leverAPIURL      string
	leverEUAPIURL    string
	greenhouseKeys   map[string]string
	leverKeys        map[string]string
	atsTimeout       time.Duration
	atsRateLimit     RateLimitPolicy
}

var applicationConfigI applicationConfig

func loadApplicationConfig() {
	applicationConfigI = applicationConfig{
		greenhouseAPIURL: strings.TrimRight(
			getEnvWithDefault(
				"GREENHOUSE_BOARDS_API_URL",
				"https://boards-api.greenhouse.io/v1",
			),
			"/",
		),
		leverAPIURL: strings.TrimRight(
			getEnvWithDefault("LEVER_POSTINGS_API_URL", "https://api.lever.co/v0"),
			"/",
		),
		leverEUAPIURL: strings.TrimRight(
			getEnvWithDefault("LEVER_EU_POSTINGS_API_URL", "https://api.eu.lever.co/v0"),
			"/",
		),
		greenhouseKeys: parseKeyPairs("GREENHOUSE_BOARD_KEYS"),
		leverKeys:      parseKeyPairs("LEVER_SITE_KEYS"),
		atsTimeout:     getEnvDurationWithDefault("ATS_HTTP_TIMEOUT", 15*time.Second),
		atsRateLimit: RateLimitPolicy{
			Name:   RateLimitPolicyName(atsRateLimitName),
			Limit:  max(getEnvIntWithDefault("ATS_RATE_LIMIT", 30), 1),
			Period: time.Minute,
			Burst:  max(getEnvIntWithDefault("ATS_RATE_BURST", 5), 1),
		},
	}
}

func parseKeyPairs(envName string) map[string]string {
	pairs := map[string]string{}

	for _, entry := range getEnvListWithDefault(envName, nil) {
		name, key, found := strings.Cut(entry, keyPairSeparator)
		if !found || name == "" || key == "" {
			log.Fatalf("%s entries must look like <board or site>:<api key>", envName)
		}

		pairs[strings.ToLower(name)] = key
	}

	return pairs
}

func GetApplicationIDParam() string             { return applicationIDParam }
func GetApplicationStepSubmitting() string      { return applicationStepSubmitting }
func GetApplicationStepSubmitted() string       { return applicationStepSubmitted }
func GetApplicationProfileKey() string          { return applicationProfileKey }
func GetGreenhouseQuestionsQuery() string       { return greenhouseQuestionsQuery }
func GetLeverSiteKeyQuery() string              { return leverSiteKeyQuery }
func GetAnswerYes() string                      { return answerYes }
func GetAnswerNo() string                       { return answerNo }
func GetFormHasHiddenQuestionsNote() string     { return formHasHiddenQuestionsNote }
func GetUnsupportedATSNote() string             { return unsupportedATSNote }
func GetGreenhouseBoardsAPIURL() string         { return applicationConfigI.greenhouseAPIURL }
func GetLeverPostingsAPIURL() string            { return applicationConfigI.leverAPIURL }
func GetGreenhouseBoardKeys() map[string]string { return applicationConfigI.greenhouseKeys }
func GetLeverSiteKeys() map[string]string       { return applicationConfigI.leverKeys }
func GetATSTimeout() time.Duration              { return applicationConfigI.atsTimeout }
func GetATSRateLimitPolicy() RateLimitPolicy    { return applicationConfigI.atsRateLimit }

func ProfileFieldForLabel(label string) string {
	lowered := strings.ToLower(label)

	for _, entry := range profileFieldKeywordsI {
		for _, keyword := range entry.Keywords {
			if strings.Contains(lowered, keyword) {
				return entry.Field
			}
		}
	}

	return ""
}
func GetLeverEUPostingsAPIURL() string  { return applicationConfigI.leverEUAPIURL }
func GetGreenhouseHostSuffix() string   { return greenhouseHostSuffix }
func GetLeverHostSuffix() string        { return leverHostSuffix }
func GetLeverEUHostMarker() string      { return leverEUHostMarker }
func GetGreenhouseJobsSegment() string  { return greenhouseJobsSegment }
func GetLeverApplySegment() string      { return leverApplySegment }
func GetLeverLinkedinURLField() string  { return leverLinkedinURLField }
func GetLeverGithubURLField() string    { return leverGithubURLField }
func GetLeverPortfolioURLField() string { return leverPortfolioURLField }
