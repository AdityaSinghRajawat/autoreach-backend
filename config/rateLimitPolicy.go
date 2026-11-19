package config

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type (
	RateLimitPolicyName string
	RateLimitKey        string
)

type RateLimitPolicy struct {
	Name            RateLimitPolicyName
	Limit           int
	Period          time.Duration
	Burst           int
	Key             RateLimitKey
	FailClosed      bool
	CountOnlyStatus int
}

const (
	RateLimitKeyIP        RateLimitKey = "ip"
	RateLimitKeyPrincipal RateLimitKey = "principal"
	RateLimitKeyAccount   RateLimitKey = "account"

	RateLimitPolicyIP                RateLimitPolicyName = "ip"
	RateLimitPolicyAuth              RateLimitPolicyName = "auth"
	RateLimitPolicyLoginAccount      RateLimitPolicyName = "login_account"
	RateLimitPolicyUser              RateLimitPolicyName = "user"
	RateLimitPolicyAPIClient         RateLimitPolicyName = "api_client"
	RateLimitPolicyResumeUpload      RateLimitPolicyName = "resume_upload"
	RateLimitPolicyContactEnrich     RateLimitPolicyName = "contact_enrich"
	RateLimitPolicyMailboxVerify     RateLimitPolicyName = "mailbox_verify"
	RateLimitPolicyOutreachCreate    RateLimitPolicyName = "outreach_create"
	RateLimitPolicyApplicationCreate RateLimitPolicyName = "application_create"
	RateLimitPolicyPublicLink        RateLimitPolicyName = "public_link"
	RateLimitPolicyAI                RateLimitPolicyName = "ai"

	rateLimitEnvPrefix = "RATE_LIMIT_"
	day                = 24 * time.Hour
)

var rateLimitDefaultsI = []RateLimitPolicy{
	{Name: RateLimitPolicyAI, Limit: 0, Period: time.Minute, Burst: 1},
	{Name: RateLimitPolicyIP, Limit: 1200, Period: time.Minute, Burst: 200, Key: RateLimitKeyIP},
	// Generous because mail providers fetch tracking pixels for many recipients from shared IPs.
	{
		Name:   RateLimitPolicyPublicLink,
		Limit:  600,
		Period: time.Minute,
		Burst:  200,
		Key:    RateLimitKeyIP,
	},
	{
		Name:       RateLimitPolicyAuth,
		Limit:      10,
		Period:     time.Minute,
		Burst:      10,
		Key:        RateLimitKeyIP,
		FailClosed: true,
	},
	{
		Name:            RateLimitPolicyLoginAccount,
		Limit:           10,
		Period:          time.Hour,
		Burst:           5,
		Key:             RateLimitKeyAccount,
		FailClosed:      true,
		CountOnlyStatus: http.StatusUnauthorized,
	},
	{
		Name:   RateLimitPolicyUser,
		Limit:  600,
		Period: time.Minute,
		Burst:  100,
		Key:    RateLimitKeyPrincipal,
	},
	{
		Name:   RateLimitPolicyAPIClient,
		Limit:  1200,
		Period: time.Minute,
		Burst:  300,
		Key:    RateLimitKeyPrincipal,
	},
	{
		Name:       RateLimitPolicyResumeUpload,
		Limit:      20,
		Period:     day,
		Burst:      10,
		Key:        RateLimitKeyPrincipal,
		FailClosed: true,
	},
	{
		Name:       RateLimitPolicyContactEnrich,
		Limit:      30,
		Period:     day,
		Burst:      10,
		Key:        RateLimitKeyPrincipal,
		FailClosed: true,
	},
	{
		Name:       RateLimitPolicyMailboxVerify,
		Limit:      20,
		Period:     time.Hour,
		Burst:      5,
		Key:        RateLimitKeyPrincipal,
		FailClosed: true,
	},
	{
		Name:       RateLimitPolicyOutreachCreate,
		Limit:      60,
		Period:     day,
		Burst:      20,
		Key:        RateLimitKeyPrincipal,
		FailClosed: true,
	},
	{
		Name:       RateLimitPolicyApplicationCreate,
		Limit:      50,
		Period:     day,
		Burst:      20,
		Key:        RateLimitKeyPrincipal,
		FailClosed: true,
	},
}

type rateLimitConfig struct {
	policies        map[RateLimitPolicyName]RateLimitPolicy
	trustedProxies  []string
	trustedPlatform string
}

var rateLimitConfigI rateLimitConfig

func loadRateLimitConfig() {
	policies := make(map[RateLimitPolicyName]RateLimitPolicy, len(rateLimitDefaultsI))

	for _, policy := range rateLimitDefaultsI {
		envName := rateLimitEnvPrefix + strings.ToUpper(string(policy.Name))
		policy.Limit = getEnvIntWithDefault(envName, policy.Limit)
		policy.Period = getEnvDurationWithDefault(envName+"_PERIOD", policy.Period)
		policy.Burst = getEnvIntWithDefault(envName+"_BURST", policy.Burst)

		if policy.Limit < 0 || (policy.Limit > 0 && (policy.Period <= 0 || policy.Burst < 1)) {
			log.Fatalf(
				"%s needs a limit of 0 or more, a positive period and a burst of 1 or more",
				envName,
			)
		}

		policies[policy.Name] = policy
	}

	trustedProxies := getEnvListWithDefault("TRUSTED_PROXIES", nil)
	for _, proxy := range trustedProxies {
		if _, _, err := net.ParseCIDR(proxy); err != nil && net.ParseIP(proxy) == nil {
			log.Fatalf("TRUSTED_PROXIES entry %q is not an IP address or CIDR range", proxy)
		}
	}

	rateLimitConfigI = rateLimitConfig{
		policies:        policies,
		trustedProxies:  trustedProxies,
		trustedPlatform: getEnvWithDefault("TRUSTED_PLATFORM", ""),
	}
}

func GetRateLimitPolicy(name RateLimitPolicyName) RateLimitPolicy {
	return rateLimitConfigI.policies[name]
}

func (p RateLimitPolicy) IsEnabled() bool { return p.Limit > 0 }

func GetTrustedProxies() []string { return rateLimitConfigI.trustedProxies }
func GetTrustedPlatform() string  { return rateLimitConfigI.trustedPlatform }
