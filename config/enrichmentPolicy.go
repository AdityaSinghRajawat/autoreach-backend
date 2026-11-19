package config

import (
	"strings"
	"time"
)

const (
	ProviderApollo     = "APOLLO"
	ProviderSignalHire = "SIGNALHIRE"
	ProviderContactOut = "CONTACTOUT"
	ProviderHunter     = "HUNTER"

	creditCostPerLookup       = 1
	signalHireWebhookPath     = "/webhooks/signalhire"
	webhookTokenParam         = "token"
	signalHireRequestIDHeader = "Request-Id"
	enrichmentRateNamespace   = "enrichment"
	enrichmentCreditNamespace = "enrichment_credits"
)

type ProviderPolicy struct {
	Code         string
	APIKey       string
	BaseURL      string
	Timeout      time.Duration
	RateLimit    int
	RateWindow   time.Duration
	RateBurst    int
	DailyCredits int

	// People returned by one discovery call; 0 means ENRICHMENT_PEOPLE_PER_SEARCH.
	PeoplePerSearch int
}

type enrichmentConfig struct {
	providerOrder           []string
	policies                map[string]ProviderPolicy
	maxContactsPerJob       int
	peoplePerSearch         int
	contactFreshness        time.Duration
	breakerWindow           time.Duration
	breakerCooldown         time.Duration
	breakerMinRequests      int
	breakerFailureRate      float64
	mxLookupTimeout         time.Duration
	publicBaseURL           string
	signalHireWebhookSecret string
}

const enrichmentSaveReserve = 5 * time.Second

var (
	enrichmentConfigI    enrichmentConfig
	providerDefaultURLsI = map[string]string{
		ProviderApollo:     "https://api.apollo.io/api/v1",
		ProviderSignalHire: "https://www.signalhire.com/api/v1",
		ProviderContactOut: "https://api.contactout.com/v1",
		ProviderHunter:     "https://api.hunter.io/v2",
	}
	providerDefaultDailyCreditsI    = map[string]int{ProviderHunter: 5}
	providerDefaultPeoplePerSearchI = map[string]int{ProviderHunter: 3}
	providerDiscoveryCreditsI       = map[string]int{ProviderHunter: 1}
	providerDomainLookupCreditsI    = map[string]int{ProviderHunter: 0}
)

const (
	defaultProviderDailyCredits = 200
	defaultProviderRateLimit    = 50
	defaultProviderRateBurst    = 10
)

// Hunter allows about 500 requests a minute, and a poll resolves one domain per company.
var (
	providerDefaultRateLimitI = map[string]int{ProviderHunter: 400}
	providerDefaultRateBurstI = map[string]int{ProviderHunter: 20}
)

func rateLimitDefault(code string) int {
	if limit, found := providerDefaultRateLimitI[code]; found {
		return limit
	}

	return defaultProviderRateLimit
}

func rateBurstDefault(code string) int {
	if burst, found := providerDefaultRateBurstI[code]; found {
		return burst
	}

	return defaultProviderRateBurst
}

func loadEnrichmentConfig() {
	timeout := getEnvDurationWithDefault("ENRICHMENT_HTTP_TIMEOUT", 8*time.Second)
	policies := make(map[string]ProviderPolicy, len(providerDefaultURLsI))

	for code, defaultURL := range providerDefaultURLsI {
		policies[code] = ProviderPolicy{
			Code:         code,
			APIKey:       getEnvWithDefault(code+"_API_KEY", ""),
			BaseURL:      getEnvWithDefault(code+"_BASE_URL", defaultURL),
			Timeout:      timeout,
			RateLimit:    getEnvIntWithDefault(code+"_RATE_LIMIT", rateLimitDefault(code)),
			RateWindow:   getEnvDurationWithDefault(code+"_RATE_WINDOW", time.Minute),
			RateBurst:    max(getEnvIntWithDefault(code+"_RATE_BURST", rateBurstDefault(code)), 1),
			DailyCredits: getEnvIntWithDefault(code+"_DAILY_CREDITS", dailyCreditsDefault(code)),
			PeoplePerSearch: getEnvIntWithDefault(
				code+"_PEOPLE_PER_SEARCH",
				providerDefaultPeoplePerSearchI[code],
			),
		}
	}

	order := getEnvListWithDefault(
		"ENRICHMENT_PROVIDERS",
		[]string{ProviderHunter, ProviderApollo, ProviderSignalHire, ProviderContactOut},
	)

	known := make([]string, 0, len(order))
	for _, code := range order {
		if _, found := policies[strings.ToUpper(code)]; found {
			known = append(known, strings.ToUpper(code))
		}
	}

	enrichmentConfigI = enrichmentConfig{
		providerOrder:     known,
		policies:          policies,
		maxContactsPerJob: max(getEnvIntWithDefault("ENRICHMENT_MAX_CONTACTS_PER_JOB", 3), 1),
		peoplePerSearch:   max(getEnvIntWithDefault("ENRICHMENT_PEOPLE_PER_SEARCH", 10), 1),
		contactFreshness: getEnvDurationWithDefault(
			"ENRICHMENT_CONTACT_FRESHNESS",
			30*24*time.Hour,
		),
		breakerWindow: getEnvDurationWithDefault(
			"ENRICHMENT_BREAKER_WINDOW",
			time.Minute,
		),
		breakerCooldown: getEnvDurationWithDefault(
			"ENRICHMENT_BREAKER_COOLDOWN",
			time.Minute,
		),
		breakerMinRequests: max(
			getEnvIntWithDefault("ENRICHMENT_BREAKER_MIN_REQUESTS", 20),
			1,
		),
		breakerFailureRate: getEnvFloatWithDefault("ENRICHMENT_BREAKER_FAILURE_RATE", 0.05),
		mxLookupTimeout: getEnvDurationWithDefault(
			"ENRICHMENT_MX_LOOKUP_TIMEOUT",
			3*time.Second,
		),
		publicBaseURL:           strings.TrimRight(getEnvWithDefault("PUBLIC_BASE_URL", ""), "/"),
		signalHireWebhookSecret: getEnvWithDefault("SIGNALHIRE_WEBHOOK_SECRET", ""),
	}
}

func GetEnrichmentProviderPolicies() []ProviderPolicy {
	policies := make([]ProviderPolicy, 0, len(enrichmentConfigI.providerOrder))
	for _, code := range enrichmentConfigI.providerOrder {
		policies = append(policies, enrichmentConfigI.policies[code])
	}

	return policies
}

func GetEnrichmentProviderPolicy(code string) (ProviderPolicy, bool) {
	policy, found := enrichmentConfigI.policies[code]

	return policy, found
}

func GetEnrichmentMaxContactsPerJob() int          { return enrichmentConfigI.maxContactsPerJob }
func GetEnrichmentPeoplePerSearch() int            { return enrichmentConfigI.peoplePerSearch }
func GetEnrichmentContactFreshness() time.Duration { return enrichmentConfigI.contactFreshness }
func GetBreakerWindow() time.Duration              { return enrichmentConfigI.breakerWindow }
func GetBreakerCooldown() time.Duration            { return enrichmentConfigI.breakerCooldown }
func GetBreakerMinRequests() int                   { return enrichmentConfigI.breakerMinRequests }
func GetBreakerFailureRate() float64               { return enrichmentConfigI.breakerFailureRate }
func GetMXLookupTimeout() time.Duration            { return enrichmentConfigI.mxLookupTimeout }

func GetSignalHireWebhookSecret() string   { return enrichmentConfigI.signalHireWebhookSecret }
func GetSignalHireWebhookPath() string     { return signalHireWebhookPath }
func GetWebhookTokenParam() string         { return webhookTokenParam }
func GetSignalHireRequestIDHeader() string { return signalHireRequestIDHeader }
func GetEnrichmentCreditNamespace() string { return enrichmentCreditNamespace }
func GetCreditCostPerLookup() int          { return creditCostPerLookup }

func GetSignalHireCallbackURL() string {
	if enrichmentConfigI.publicBaseURL == "" || enrichmentConfigI.signalHireWebhookSecret == "" {
		return ""
	}

	return enrichmentConfigI.publicBaseURL + constI.http.apiBasePath + signalHireWebhookPath +
		"?" + webhookTokenParam + "=" + enrichmentConfigI.signalHireWebhookSecret
}

func dailyCreditsDefault(code string) int {
	if credits, found := providerDefaultDailyCreditsI[code]; found {
		return credits
	}

	return defaultProviderDailyCredits
}

func (p ProviderPolicy) DiscoveryLimit(requested int) int {
	if p.PeoplePerSearch > 0 {
		return min(requested, p.PeoplePerSearch)
	}

	return requested
}

// Providers that charge per person found while searching (Hunter) pay up front for the worst case.
func GetDiscoveryCredits(code string, people int) int {
	return providerDiscoveryCreditsI[code] * people
}

// Hunter's Domain Finder is free; other providers charge the standard lookup price.
func GetDomainLookupCredits(code string) int {
	if credits, found := providerDomainLookupCreditsI[code]; found {
		return credits
	}

	return companyDomainLookupCredits
}

func (p ProviderPolicy) RateLimitPolicy() RateLimitPolicy {
	return RateLimitPolicy{
		Name:   RateLimitPolicyName(enrichmentRateNamespace),
		Limit:  p.RateLimit,
		Period: p.RateWindow,
		Burst:  p.RateBurst,
	}
}

func GetPublicBaseURL() string { return enrichmentConfigI.publicBaseURL }

// Task time kept for saving contacts already paid for.
func GetEnrichmentSaveReserve() time.Duration { return enrichmentSaveReserve }

// Legal suffixes left out of company-name searches, since profiles often omit them.
var companyLegalSuffixesI = map[string]struct{}{
	"inc": {}, "llc": {}, "ltd": {}, "limited": {}, "corp": {}, "corporation": {}, "co": {},
	"company": {}, "gmbh": {}, "plc": {}, "pvt": {}, "private": {}, "sa": {}, "ag": {}, "bv": {},
}

// Words that are operators in a provider's boolean query syntax.
var booleanOperatorsI = map[string]struct{}{"and": {}, "or": {}, "not": {}}

func IsCompanyLegalSuffix(word string) bool {
	_, suffix := companyLegalSuffixesI[strings.ToLower(word)]

	return suffix
}

func IsBooleanOperator(word string) bool {
	_, operator := booleanOperatorsI[strings.ToLower(word)]

	return operator
}
