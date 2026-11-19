package unit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestContactTitlesAreClassifiedIntoRoles(t *testing.T) {
	cases := map[string]string{
		"Engineering Manager, Platform":  config.ContactRoleEngineeringManager,
		"VP of Engineering":              config.ContactRoleEngineeringManager,
		"Tech Lead":                      config.ContactRoleEngineeringManager,
		"Senior Technical Recruiter":     config.ContactRoleRecruiter,
		"Head of Talent Acquisition":     config.ContactRoleRecruiter,
		"Engineering Recruiting Manager": config.ContactRoleRecruiter,
		"Director of Marketing":          config.ContactRoleHiringManager,
		"Head of Design":                 config.ContactRoleHiringManager,
	}
	for title, expected := range cases {
		role, classified := config.ClassifyContactRole(title)
		assert.True(t, classified, "title %q", title)
		assert.Equal(t, expected, role, "title %q", title)
	}

	_, classified := config.ClassifyContactRole("Software Engineer II")
	assert.False(t, classified, "an individual contributor is not a decision maker")
}

func TestTargetRolesDependOnWhetherTheJobIsAnEngineeringRole(t *testing.T) {
	assert.Equal(
		t,
		[]string{config.ContactRoleEngineeringManager, config.ContactRoleRecruiter},
		config.GetDefaultTargetRoles("Senior Go Engineer", ""),
	)
	assert.Equal(
		t,
		[]string{config.ContactRoleHiringManager, config.ContactRoleRecruiter},
		config.GetDefaultTargetRoles("Account Executive", "Sales"),
	)

	job := jobModel.NewJobPosting()
	job.JobTitle = "Account Executive"

	assert.Equal(
		t,
		[]string{config.ContactRoleRecruiter},
		enrichmentTypes.ResolveTargetRoles([]string{config.ContactRoleRecruiter}, job),
		"explicit roles win over the default",
	)
}

func TestEmailVerificationCombinesSyntaxDNSRoleAndProviderSignals(t *testing.T) {
	setup.Config(t)

	verifier := utils.NewEmailVerifier(helpers.FakeMXResolver{
		Domains: map[string]bool{"acme.com": true, "mailinator.com": true},
	})
	ctx := context.Background()

	assert.Equal(t, modelHelpers.EmailStatusVerified, verifier.Verify(ctx, "sarah@acme.com", true))
	assert.Equal(t, modelHelpers.EmailStatusRisky, verifier.Verify(ctx, "sarah@acme.com", false),
		"deliverable but not confirmed by the provider")
	assert.Equal(t, modelHelpers.EmailStatusRisky, verifier.Verify(ctx, "careers@acme.com", true),
		"a shared mailbox is not a person")
	assert.Equal(
		t,
		modelHelpers.EmailStatusInvalid,
		verifier.Verify(ctx, "sarah@nomx.example", true),
	)
	assert.Equal(
		t,
		modelHelpers.EmailStatusInvalid,
		verifier.Verify(ctx, "sarah@mailinator.com", true),
	)

	for _, malformed := range []string{"", "not-an-email", "sarah@", "@acme.com", "Sarah <sarah@acme.com>", "sarah@localhost"} {
		assert.Equal(
			t,
			modelHelpers.EmailStatusInvalid,
			verifier.Verify(ctx, malformed, true),
			"email %q",
			malformed,
		)
	}
}

func TestCircuitBreakerTripsOnRateLimitAndOnFailureRateThenRecovers(t *testing.T) {
	setup.Config(t)

	rateLimited := utils.NewCircuitBreaker()
	require.True(t, rateLimited.Allow())
	rateLimited.Record(false, true)
	assert.False(t, rateLimited.Allow(), "a 429 opens the circuit at once")

	time.Sleep(config.GetBreakerCooldown() + 50*time.Millisecond)
	assert.True(t, rateLimited.Allow(), "the circuit closes again after the cooldown")

	failing := utils.NewCircuitBreaker()
	for range config.GetBreakerMinRequests() - 1 {
		failing.Record(true, false)
	}

	assert.True(t, failing.Allow(), "too few samples to judge")
	failing.Record(true, false)
	assert.False(t, failing.Allow(), "the failure rate is over the threshold")

	healthy := utils.NewCircuitBreaker()
	for range config.GetBreakerMinRequests() * 3 {
		healthy.Record(false, false)
	}

	assert.True(t, healthy.Allow())
}

func TestCircuitBreakerLetsOneProbeThroughAfterTheCooldown(t *testing.T) {
	setup.Config(t)

	breaker := utils.NewCircuitBreaker()
	breaker.Record(false, true)
	time.Sleep(config.GetBreakerCooldown() + 50*time.Millisecond)

	require.True(t, breaker.Allow(), "one probe is let through")
	assert.False(t, breaker.Allow(), "everything else waits for the probe's result")

	breaker.Record(true, false)
	assert.False(t, breaker.Allow(), "a failed probe opens the circuit again")

	time.Sleep(config.GetBreakerCooldown() + 50*time.Millisecond)
	require.True(t, breaker.Allow())
	breaker.Record(false, false)
	assert.True(t, breaker.Allow(), "a successful probe closes the circuit")
	assert.True(t, breaker.Allow())
}

func TestProviderRegistrySkipsProvidersWithoutCredentialsOrCallback(t *testing.T) {
	stub := helpers.NewEnrichmentStub(t)
	policies := stub.Policies()

	full := providers.NewProviderRegistry(policies, "https://api.autoreach.test/callback")
	require.Len(t, full, 3)
	assert.Equal(t, config.ProviderApollo, full[0].Code())
	assert.Equal(t, config.ProviderSignalHire, full[1].Code())
	assert.Equal(t, config.ProviderContactOut, full[2].Code())

	withoutCallback := providers.NewProviderRegistry(policies, "")
	require.Len(t, withoutCallback, 2, "SignalHire needs a callback URL")

	policies[0].APIKey = ""
	withoutApolloKey := providers.NewProviderRegistry(
		policies,
		"https://api.autoreach.test/callback",
	)
	require.Len(t, withoutApolloKey, 2)
	assert.Equal(t, config.ProviderSignalHire, withoutApolloKey[0].Code())
}

func TestProviderAdaptersTranslateVendorPayloads(t *testing.T) {
	stub := helpers.NewEnrichmentStub(t)
	stub.People = []map[string]any{
		{
			"id":           "p1",
			"first_name":   "Sarah",
			"last_name":    "Li",
			"title":        "Engineering Manager",
			"linkedin_url": "http://www.linkedin.com/in/sarah-li",
			"organization": map[string]any{"name": "Acme Corp"},
		},
		{"id": "", "title": "No Identifier"},
	}
	stub.Matches["p1"] = map[string]any{
		"id": "p1", "name": "Sarah Li", "title": "Engineering Manager", "email": "sarah@acme.com",
		"email_status": "verified", "linkedin_url": "http://www.linkedin.com/in/sarah-li",
	}
	stub.Profiles["https://www.linkedin.com/in/sarah-li"] = map[string]any{
		"url": "https://www.linkedin.com/in/sarah-li", "full_name": "Sarah Li",
		"work_email": []string{"s.li@acme.com"}, "phone": []string{"+15550100"},
	}

	registry := providers.NewProviderRegistry(
		stub.Policies(),
		"https://api.autoreach.test/callback",
	)
	ctx := context.Background()

	people, err := registry[0].(providers.Discoverer).DiscoverPeople(
		ctx,
		enrichmentTypes.DiscoveryQuery{
			CompanyDomain: "acme.com", Titles: []string{"engineering manager"}, Limit: 5,
		},
	)
	require.NoError(t, err)
	require.Len(t, people, 1, "a person without an id cannot be enriched and is dropped")
	assert.Equal(t, "Sarah Li", people[0].FullName)
	assert.Equal(t, "Acme Corp", people[0].CompanyName)

	apolloDetails, err := registry[0].(providers.ContactFinder).FindContact(ctx, people[0])
	require.NoError(t, err)
	assert.Equal(t, "sarah@acme.com", apolloDetails.Email)
	assert.True(t, apolloDetails.EmailVerified)

	person := people[0]
	person.LinkedinURL = "https://www.linkedin.com/in/sarah-li"

	contactOutDetails, err := registry[2].(providers.ContactFinder).FindContact(ctx, person)
	require.NoError(t, err)
	assert.Equal(t, "s.li@acme.com", contactOutDetails.Email)
	assert.Equal(t, "+15550100", contactOutDetails.Phone)

	person.LinkedinURL = "https://www.linkedin.com/in/unknown"
	missing, err := registry[2].(providers.ContactFinder).FindContact(ctx, person)
	require.NoError(t, err)
	assert.Nil(t, missing, "an empty profile from the provider means not found, not a failure")

	asyncFinder := registry[1].(providers.AsyncContactFinder)

	requestID, err := asyncFinder.RequestContact(ctx, people[0])
	require.NoError(t, err)
	assert.Equal(t, "1", requestID)
	assert.Equal(t, []string{"http://www.linkedin.com/in/sarah-li"}, stub.SignalHireItems())

	results, err := asyncFinder.ParseCallback([]byte(`[
		{"item":"x","status":"failed"},
		{"item":"y","status":"success","candidate":{"fullName":"Sarah Li","contacts":[
			{"type":"email","value":"personal@gmail.com","rating":"100","subType":"personal"},
			{"type":"email","value":"sarah@acme.com","rating":"100","subType":"work"},
			{"type":"phone","value":"+15550100"}]}}
	]`))
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.False(t, results[0].Found)
	assert.Equal(t, "sarah@acme.com", results[1].Details.Email, "the work address is preferred")
	assert.True(t, results[1].Details.EmailVerified)

	_, err = asyncFinder.ParseCallback([]byte(`{"not":"an array"}`))
	require.ErrorIs(t, err, config.ErrMalformedProviderPayload)

	stub.Set(func(stub *helpers.EnrichmentStub) { stub.SearchStatus = http.StatusTooManyRequests })

	_, err = registry[0].(providers.Discoverer).DiscoverPeople(
		ctx,
		enrichmentTypes.DiscoveryQuery{CompanyDomain: "acme.com"},
	)

	var apiErr *utils.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.IsTransient())
}

func TestPersonConvertsToACompanyContact(t *testing.T) {
	setup.Config(t)

	job := jobModel.NewJobPosting()
	job.CompanyDomain = "acme.com"
	job.CompanyName = "Acme Corp"

	person := enrichmentTypes.Person{
		ProviderPersonID: "p1",
		FullName:         "S. L.",
		Title:            "Engineering Manager",
		LinkedinURL:      "http://www.linkedin.com/in/sarah-li/",
	}
	details := &enrichmentTypes.ContactDetails{
		FullName: "Sarah Li", Email: "Sarah@Acme.com", EmailVerified: true, Phone: "+15550100",
		Raw: []byte(`{"id":"p1"}`),
	}

	contact := person.ToCompanyContactModel(
		job, config.ContactRoleEngineeringManager, config.ProviderApollo, details,
		modelHelpers.EmailStatusVerified,
	)

	assert.Equal(t, "Sarah Li", contact.FullName, "the enriched name replaces the obfuscated one")
	assert.Equal(t, "acme.com", contact.CompanyDomain)
	require.NotNil(t, contact.WorkEmail)
	assert.Equal(t, "sarah@acme.com", *contact.WorkEmail)
	assert.True(t, contact.IsEmailVerified)
	require.NotNil(t, contact.LinkedinURL)
	assert.Equal(t, "https://www.linkedin.com/in/sarah-li", *contact.LinkedinURL)
	assert.JSONEq(t, `{"id":"p1"}`, string(contact.EnrichmentMetadata))
	assert.True(t, contact.IsReachable())
	require.NoError(t, contact.Validate())

	invalidEmail := person.ToCompanyContactModel(
		job, config.ContactRoleEngineeringManager, config.ProviderApollo, details,
		modelHelpers.EmailStatusInvalid,
	)
	assert.Nil(t, invalidEmail.WorkEmail, "an invalid address is not stored")
	assert.True(t, invalidEmail.IsReachable(), "still reachable through LinkedIn")

	nameless := enrichmentTypes.Person{ProviderPersonID: "p2", Title: "Recruiter"}
	require.Error(
		t,
		nameless.ToCompanyContactModel(
			job, config.ContactRoleRecruiter, config.ProviderApollo, nil,
			modelHelpers.EmailStatusUnverified,
		).Validate(),
	)
}

func TestHunterReturnsPeopleWithTheirEmailsAndSkipsIncompleteOnes(t *testing.T) {
	stub := helpers.NewEnrichmentStub(t)
	stub.HunterEmails = []map[string]any{
		{
			"value": "Sarah.Li@Acme.com", "first_name": "Sarah", "last_name": "Li",
			"position": "Engineering Manager", "linkedin": "sarah-li",
			"verification": map[string]any{"status": "valid"},
		},
		{
			"value": "omar@acme.com", "first_name": "Omar", "last_name": "Khan",
			"position": "Recruiter", "linkedin": "https://www.linkedin.com/in/omar-khan/",
			"verification": map[string]any{"status": "accept_all"},
		},
		{
			"value":      "lee@acme.com",
			"first_name": "Lee",
			"last_name":  "Park",
			"position":   "Hiring Manager",
			"linkedin":   "not a handle!",
		},
		{"value": "dana@acme.com", "first_name": "Dana", "last_name": "", "position": "Director"},
	}

	registry := providers.NewProviderRegistry([]config.ProviderPolicy{stub.HunterPolicy()}, "")
	require.Len(t, registry, 1)

	discoverer, isDiscoverer := registry[0].(providers.Discoverer)
	require.True(t, isDiscoverer)
	assert.NotImplements(t, (*providers.ContactFinder)(nil), registry[0])

	people, err := discoverer.DiscoverPeople(context.Background(), enrichmentTypes.DiscoveryQuery{
		CompanyDomain: "acme.com",
		CompanyName:   "Acme",
		Titles:        []string{"recruiter", "engineering manager"},
		Limit:         3,
	})

	require.NoError(t, err)
	require.Len(t, people, 3, "a person without a full name cannot be contacted")
	assert.Equal(t, config.ProviderHunter, people[0].Source)
	assert.Equal(t, "sarah.li@acme.com", people[0].Email)
	assert.Equal(t, people[0].Email, people[0].ProviderPersonID)
	assert.True(t, people[0].EmailVerified)
	assert.Equal(t, "https://www.linkedin.com/in/sarah-li", people[0].LinkedinURL)
	assert.False(
		t,
		people[1].EmailVerified,
		"an accept-all domain does not prove the mailbox exists",
	)
	assert.Equal(t, "https://www.linkedin.com/in/omar-khan", people[1].LinkedinURL)
	assert.Empty(t, people[2].LinkedinURL, "a value that is not a profile handle is dropped")

	queries := stub.HunterQueries()
	require.Len(t, queries, 1)
	assert.Equal(t, "acme.com", queries[0].Get("domain"))
	assert.Equal(t, "recruiter,engineering manager", queries[0].Get("job_titles"))
	assert.Equal(t, "3", queries[0].Get("limit"))
	assert.Equal(t, "personal", queries[0].Get("type"))

	stub.Set(
		func(stub *helpers.EnrichmentStub) { stub.HunterStatus = http.StatusUnavailableForLegalReasons },
	)

	none, err := discoverer.DiscoverPeople(
		context.Background(),
		enrichmentTypes.DiscoveryQuery{CompanyDomain: "acme.com", Limit: 3},
	)
	require.NoError(t, err)
	assert.Empty(t, none, "people who asked not to be processed are never returned")

	stub.Set(func(stub *helpers.EnrichmentStub) { stub.HunterStatus = http.StatusTooManyRequests })

	_, err = discoverer.DiscoverPeople(
		context.Background(),
		enrichmentTypes.DiscoveryQuery{CompanyDomain: "acme.com", Limit: 3},
	)

	var apiErr *utils.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.IsTransient())
}

func TestPaidDiscoveryIsCappedPerSearchAndReservesItsCredits(t *testing.T) {
	capped := config.ProviderPolicy{PeoplePerSearch: 3}
	assert.Equal(t, 3, capped.DiscoveryLimit(10))
	assert.Equal(t, 2, capped.DiscoveryLimit(2))
	assert.Equal(
		t,
		10,
		config.ProviderPolicy{}.DiscoveryLimit(10),
		"no cap means the global search size",
	)

	assert.Equal(
		t,
		3,
		config.GetDiscoveryCredits(config.ProviderHunter, 3),
		"Hunter charges for each person found",
	)
	assert.Zero(
		t,
		config.GetDiscoveryCredits(config.ProviderApollo, 10),
		"other providers charge when the contact is looked up",
	)
}

func TestHunterResolvesACompanyDomainOnlyForAnExactNameMatchAtNoCost(t *testing.T) {
	stub := helpers.NewEnrichmentStub(t)
	stub.HunterDomains = map[string][]map[string]any{
		"Acme Staffing Pvt Ltd": {
			{"domain": "acmefitness.com", "company_name": "ACME Fitness Pvt"},
			{"domain": "acmestaffing.com", "company_name": "ACME Staffing"},
		},
		"Gojek Tech":           {{"domain": "gojek.com", "company_name": "Gojek"}},
		"Digirex Technologies": {{"domain": "akamai.com", "company_name": "Akamai Technologies"}},
	}

	registry := providers.NewProviderRegistry([]config.ProviderPolicy{stub.HunterPolicy()}, "")
	require.Len(t, registry, 1)

	resolver, isResolver := registry[0].(providers.DomainResolver)
	require.True(t, isResolver)

	domain, err := resolver.ResolveCompanyDomain(context.Background(), "Acme Staffing Pvt Ltd")
	require.NoError(t, err)
	assert.Equal(t, "acmestaffing.com", domain, "a look-alike ahead of the match is skipped")

	domain, err = resolver.ResolveCompanyDomain(context.Background(), "Gojek Tech")
	require.NoError(t, err)
	assert.Equal(t, "gojek.com", domain, "a generic trailing word does not change the company")

	domain, err = resolver.ResolveCompanyDomain(context.Background(), "Digirex Technologies")
	require.NoError(t, err)
	assert.Empty(t, domain, "a different company name is not accepted")

	domain, err = resolver.ResolveCompanyDomain(context.Background(), "pi")
	require.NoError(t, err)
	assert.Empty(t, domain, "Hunter rejects names under three characters, so none is sent")
	assert.Equal(t, 3, stub.Calls("hunter.domainFinder"), "no call was made for the short name")

	domain, err = resolver.ResolveCompanyDomain(context.Background(), "Unknown Co")
	require.NoError(t, err)
	assert.Empty(t, domain)

	assert.Zero(t, config.GetDomainLookupCredits(config.ProviderHunter), "Domain Finder is free")
	assert.Positive(t, config.GetDomainLookupCredits(config.ProviderApollo))
}

func TestCompanyNamesMatchIgnoresSpacingAndGenericWordsButNotDifferentCompanies(t *testing.T) {
	matching := [][2]string{
		{"Code Vyasa", "CodeVyasa"},
		{"Benison Technologies", "Benison"},
		{"Awign Enterprises", "Awign"},
		{"Acme, Inc.", "ACME Inc"},
		{"Extreme Networks India Private", "Extreme Networks"},
	}
	for _, pair := range matching {
		assert.True(
			t,
			modelHelpers.CompanyNamesMatch(pair[0], pair[1]),
			"%q vs %q",
			pair[0],
			pair[1],
		)
	}

	different := [][2]string{
		{"Digirex Technologies", "Akamai Technologies"},
		{"Extreme Networks India Private", "Mistral Networks India Private Limited"},
		{"Datamatics Global Services Limited", "Datamatics Staffing Services"},
		{"Technologies", "Services"},
		{"", ""},
	}
	for _, pair := range different {
		assert.False(
			t,
			modelHelpers.CompanyNamesMatch(pair[0], pair[1]),
			"%q vs %q",
			pair[0],
			pair[1],
		)
	}
}
