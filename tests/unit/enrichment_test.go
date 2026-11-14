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
	assert.Nil(t, missing, "a 404 from the provider means not found, not a failure")

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
