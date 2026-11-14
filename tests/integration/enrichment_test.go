//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

const (
	sarahLinkedin = "https://www.linkedin.com/in/sarah-li"
	omarLinkedin  = "https://www.linkedin.com/in/omar-khan"
)

type enrichmentFixture struct {
	jobFixture

	stub  *helpers.EnrichmentStub
	jobID string
}

func newEnrichmentFixture(t *testing.T) enrichmentFixture {
	t.Helper()

	fixture := newJobFixture(t)
	pinMatchingVectors(setup.Embedder(t))

	stub := helpers.NewEnrichmentStub(t)
	stub.People = []map[string]any{
		{
			"id":           "p1",
			"first_name":   "Sarah",
			"last_name":    "L.",
			"title":        "Engineering Manager",
			"linkedin_url": sarahLinkedin,
		},
		{
			"id":           "p2",
			"first_name":   "Omar",
			"last_name":    "K.",
			"title":        "Senior Technical Recruiter",
			"linkedin_url": omarLinkedin,
		},
		{"id": "p3", "first_name": "Dana", "last_name": "P.", "title": "Director of Marketing"},
		{"id": "p4", "first_name": "Ravi", "last_name": "S.", "title": "Software Engineer"},
	}
	stub.Matches["p1"] = map[string]any{
		"id": "p1", "name": "Sarah Li", "title": "Engineering Manager", "email": "sarah@acme.com",
		"email_status": "verified", "linkedin_url": sarahLinkedin,
	}
	stub.Matches["p2"] = map[string]any{
		"id":           "p2",
		"name":         "Omar Khan",
		"title":        "Senior Technical Recruiter",
		"email":        "omar@acme.com",
		"email_status": "verified",
		"linkedin_url": omarLinkedin,
	}

	fixture.deps.EnrichmentProviders = providers.NewProviderRegistry(
		stub.Policies(),
		config.GetSignalHireCallbackURL(),
	)
	fixture.deps.EmailVerifier = utils.NewEmailVerifier(helpers.FakeMXResolver{
		Domains: map[string]bool{"acme.com": true},
	})
	require.NoError(t, fixture.deps.Wire())

	fixture.router = routes.NewRoutes(fixture.deps)

	t.Cleanup(func() {
		fixture.deps.EnrichmentProviders = []providers.Provider{}
		fixture.deps.EmailVerifier = utils.NewEmailVerifier(nil)
		require.NoError(t, fixture.deps.Wire())
	})

	jobID := fixture.embeddedJob(t, "Senior Go Engineer", "We need Go. "+nearJobMarker)
	resume := fixture.processedResume(t, "Alex Candidate "+resumeMarker)

	_, customErr := fixture.deps.Services.Matching.MatchResume(context.Background(), resume.ID)
	require.Nil(t, customErr)

	return enrichmentFixture{jobFixture: fixture, stub: stub, jobID: jobID}
}

func (f enrichmentFixture) enrich(t *testing.T) *enrichmentTypes.EnrichJobContactsResult {
	t.Helper()

	result, customErr := f.deps.Services.Enrichment.EnrichJobContacts(
		context.Background(),
		f.jobID,
		nil,
	)
	require.Nil(t, customErr)

	return result
}

func (f enrichmentFixture) jobContacts(t *testing.T) []enrichmentModel.CompanyContact {
	t.Helper()

	exchange := helpers.Do(
		t, f.router, http.MethodGet,
		candidatePath(f.candidateID)+"/jobs/"+f.jobID+"/contacts", nil, bearer(f.userToken),
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[[]enrichmentModel.CompanyContact](t, exchange)
}

func contactByName(
	t *testing.T,
	contacts []enrichmentModel.CompanyContact,
	name string,
) enrichmentModel.CompanyContact {
	t.Helper()

	for _, contact := range contacts {
		if contact.FullName == name {
			return contact
		}
	}

	require.Failf(t, "contact not found", "no contact named %q in %+v", name, contacts)

	return enrichmentModel.CompanyContact{}
}

func TestEnrichmentFindsDecisionMakersThroughThePrimaryProvider(t *testing.T) {
	fixture := newEnrichmentFixture(t)

	result := fixture.enrich(t)

	assert.Equal(t, 4, result.Discovered)
	assert.Equal(
		t,
		2,
		result.Saved,
		"only the engineering manager and the recruiter fit an engineering job",
	)
	assert.Zero(t, result.PendingAsync)
	assert.Equal(t, 1, fixture.stub.Calls("apollo.search"))
	assert.Equal(t, 2, fixture.stub.Calls("apollo.match"))
	assert.Zero(
		t,
		fixture.stub.Calls("contactout.enrich"),
		"no fallback is called once an email is found",
	)

	contacts := fixture.jobContacts(t)
	require.Len(t, contacts, 2)

	sarah := contactByName(t, contacts, "Sarah Li")
	assert.EqualValues(t, config.ContactRoleEngineeringManager, sarah.RoleType)
	require.NotNil(t, sarah.WorkEmail)
	assert.Equal(t, "sarah@acme.com", *sarah.WorkEmail)
	assert.True(t, sarah.IsEmailVerified)
	assert.Equal(t, modelHelpers.EmailStatusVerified, sarah.EmailVerificationStatus)
	assert.Equal(t, config.ProviderApollo, sarah.EnrichmentProvider)
	assert.Equal(t, "acme.com", sarah.CompanyDomain)
	assert.EqualValues(
		t,
		config.ContactRoleRecruiter,
		contactByName(t, contacts, "Omar Khan").RoleType,
	)

	matched, _ := fixture.listMatchedJobs(t, "")
	require.Len(t, matched, 1)
	assert.Equal(t, 2, matched[0].DiscoveredContactsCount)

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(fixture.candidateID)+"/contacts?perPage=1",
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code, string(listed.Body))
	helpers.AssertMatchesSpec(t, listed)
	assert.Len(t, helpers.DecodeData[[]enrichmentModel.CompanyContact](t, listed), 1)
	assert.JSONEq(
		t,
		`{"page":1,"perPage":1,"totalItems":2,"totalPages":2}`,
		string(helpers.DecodeEnvelope(t, listed).Meta),
	)
}

func TestEnrichmentFallsBackToTheNextProviderWhenThePrimaryCannotFindAnEmail(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	fixture.stub.Set(func(stub *helpers.EnrichmentStub) {
		stub.MatchStatus = http.StatusInternalServerError
		stub.Profiles[sarahLinkedin] = map[string]any{
			"url": sarahLinkedin, "full_name": "Sarah Li", "work_email": []string{"s.li@acme.com"},
		}
		stub.Profiles[omarLinkedin] = map[string]any{
			"url": omarLinkedin, "full_name": "Omar Khan", "email": []string{"omar@acme.com"},
		}
	})

	result := fixture.enrich(t)

	assert.Equal(t, 2, result.Saved)
	assert.Equal(t, 2, fixture.stub.Calls("contactout.enrich"))

	contacts := fixture.jobContacts(t)
	sarah := contactByName(t, contacts, "Sarah Li")
	require.NotNil(t, sarah.WorkEmail)
	assert.Equal(t, "s.li@acme.com", *sarah.WorkEmail)
	assert.Equal(t, config.ProviderContactOut, sarah.EnrichmentProvider)
	assert.True(t, sarah.IsEmailVerified)

	omar := contactByName(t, contacts, "Omar Khan")
	assert.False(
		t,
		omar.IsEmailVerified,
		"an address the provider did not confirm as a work email is only risky",
	)
	assert.Equal(t, modelHelpers.EmailStatusRisky, omar.EmailVerificationStatus)
}

func TestRateLimitedDiscoveryOpensTheCircuitAndIsRetriedLater(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	fixture.stub.Set(
		func(stub *helpers.EnrichmentStub) { stub.SearchStatus = http.StatusTooManyRequests },
	)

	_, rateLimited := fixture.deps.Services.Enrichment.EnrichJobContacts(
		context.Background(),
		fixture.jobID,
		nil,
	)
	require.NotNil(t, rateLimited)
	assert.True(t, config.IsRetryableErrCode(rateLimited.ErrCode))
	assert.Equal(t, 1, fixture.stub.Calls("apollo.search"))

	_, whileOpen := fixture.deps.Services.Enrichment.EnrichJobContacts(
		context.Background(),
		fixture.jobID,
		nil,
	)
	require.NotNil(t, whileOpen)
	assert.ErrorIs(t, whileOpen, config.ErrProviderUnavailable)
	assert.Equal(
		t,
		1,
		fixture.stub.Calls("apollo.search"),
		"an open circuit must not call the provider",
	)

	fixture.stub.Set(func(stub *helpers.EnrichmentStub) { stub.SearchStatus = http.StatusOK })
	time.Sleep(config.GetBreakerCooldown() + 100*time.Millisecond)

	assert.Equal(t, 2, fixture.enrich(t).Saved)
}

func TestContactWithoutAnEmailIsCompletedLaterByTheSignalHireCallback(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	fixture.stub.Set(func(stub *helpers.EnrichmentStub) {
		stub.People = stub.People[:1]
		stub.Matches["p1"] = map[string]any{
			"id":           "p1",
			"name":         "Sarah Li",
			"title":        "Engineering Manager",
			"linkedin_url": sarahLinkedin,
		}
	})

	result := fixture.enrich(t)

	assert.Equal(t, 1, result.Saved)
	assert.Equal(t, 1, result.PendingAsync)
	assert.Equal(t, []string{sarahLinkedin}, fixture.stub.SignalHireItems())

	pending := contactByName(t, fixture.jobContacts(t), "Sarah Li")
	assert.Nil(t, pending.WorkEmail)
	require.NotNil(t, pending.LinkedinURL)

	callback := `[{"item":"` + sarahLinkedin + `","status":"success","candidate":{"fullName":"Sarah Li","contacts":[
		{"type":"email","value":"Sarah@Acme.com","rating":"100","subType":"work"},{"type":"phone","value":"+15550100"}]}}]`
	path := config.GetAPIBasePath() + config.GetSignalHireWebhookPath()
	requestID := map[string]string{
		config.GetSignalHireRequestIDHeader(): helpers.SignalHireRequestID(1),
	}

	forged := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		path+"?token=wrong",
		callback,
		requestID,
	)
	missingToken := helpers.Do(t, fixture.router, http.MethodPost, path, callback, requestID)
	assert.Equal(t, http.StatusUnauthorized, forged.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, missingToken.Recorder.Code)
	assert.Nil(t, contactByName(t, fixture.jobContacts(t), "Sarah Li").WorkEmail)

	authorised := path + "?token=" + config.GetSignalHireWebhookSecret()

	accepted := helpers.Do(t, fixture.router, http.MethodPost, authorised, callback, requestID)
	require.Equal(t, http.StatusOK, accepted.Recorder.Code, string(accepted.Body))
	helpers.AssertMatchesSpec(t, accepted)

	completed := contactByName(t, fixture.jobContacts(t), "Sarah Li")
	require.NotNil(t, completed.WorkEmail)
	assert.Equal(t, "sarah@acme.com", *completed.WorkEmail)
	assert.True(t, completed.IsEmailVerified)
	assert.Equal(t, config.ProviderSignalHire, completed.EnrichmentProvider)
	require.NotNil(t, completed.PhoneNumber)
	assert.Equal(t, pending.ID, completed.ID, "the callback completes the same contact")

	replayed := `[{"item":"x","status":"success","candidate":{"fullName":"Sarah Li","contacts":[{"type":"email","value":"attacker@acme.com","rating":"100","subType":"work"}]}}]`
	replay := helpers.Do(t, fixture.router, http.MethodPost, authorised, replayed, requestID)
	unknown := helpers.Do(t, fixture.router, http.MethodPost, authorised, replayed,
		map[string]string{config.GetSignalHireRequestIDHeader(): "999"})
	malformed := helpers.Do(t, fixture.router, http.MethodPost, authorised, `not json`, requestID)

	assert.Equal(t, http.StatusOK, replay.Recorder.Code)
	assert.Equal(t, http.StatusOK, unknown.Recorder.Code)
	assert.Equal(
		t,
		http.StatusOK,
		malformed.Recorder.Code,
		"a request that is no longer pending is acknowledged and ignored",
	)
	assert.Equal(
		t,
		"sarah@acme.com",
		*contactByName(t, fixture.jobContacts(t), "Sarah Li").WorkEmail,
	)
}

func TestKnownCompanyContactsAreReusedWithoutCallingAnyProvider(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	require.Equal(t, 2, fixture.enrich(t).Saved)

	secondJobID := fixture.embeddedJob(t, "Staff Go Engineer", "We need Go. "+nearJobMarker)

	result, customErr := fixture.deps.Services.Enrichment.EnrichJobContacts(
		context.Background(),
		secondJobID,
		nil,
	)
	require.Nil(t, customErr)

	assert.Equal(t, 2, result.Reused)
	assert.Zero(t, result.Saved)
	assert.Equal(
		t,
		1,
		fixture.stub.Calls("apollo.search"),
		"the second opening at the company costs nothing",
	)
	assert.Equal(t, 2, fixture.stub.Calls("apollo.match"))

	var links int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM job_contacts WHERE job_id = $1`, secondJobID,
	).Scan(&links))
	assert.Equal(t, 2, links)

	again := fixture.enrich(t)
	assert.Equal(t, 2, again.Reused, "running the same job again is a no-op")

	var contacts int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		context.Background(), `SELECT COUNT(*) FROM company_contacts`,
	).Scan(&contacts))
	assert.Equal(t, 2, contacts)
}

func TestUndeliverableAndSharedAddressesAreNotTreatedAsVerified(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	fixture.stub.Set(func(stub *helpers.EnrichmentStub) {
		stub.Matches["p1"]["email"] = "sarah@no-mail-server.example"
		stub.Matches["p2"]["email"] = "careers@acme.com"
	})

	require.Equal(t, 2, fixture.enrich(t).Saved)

	contacts := fixture.jobContacts(t)

	sarah := contactByName(t, contacts, "Sarah Li")
	assert.Nil(t, sarah.WorkEmail, "an address with no mail server is dropped")
	assert.Equal(t, modelHelpers.EmailStatusInvalid, sarah.EmailVerificationStatus)
	assert.NotNil(t, sarah.LinkedinURL, "the contact is kept because LinkedIn still reaches them")

	omar := contactByName(t, contacts, "Omar Khan")
	require.NotNil(t, omar.WorkEmail)
	assert.False(t, omar.IsEmailVerified)
	assert.Equal(t, modelHelpers.EmailStatusRisky, omar.EmailVerificationStatus)
}

func TestExhaustedCreditBudgetSkipsThePaidLookup(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	policy, _ := config.GetEnrichmentProviderPolicy(config.ProviderApollo)

	_, err := fixture.deps.RedisUtil.IncrementDaily(
		context.Background(),
		config.GetEnrichmentCreditNamespace(),
		config.ProviderApollo,
		int64(policy.DailyCredits),
	)
	require.NoError(t, err)

	result := fixture.enrich(t)

	assert.Equal(t, 1, fixture.stub.Calls("apollo.search"), "search is free and still runs")
	assert.Zero(
		t,
		fixture.stub.Calls("apollo.match"),
		"paid lookups stop once the daily budget is spent",
	)
	assert.Equal(t, 2, result.Saved, "people are still stored through their LinkedIn profile")
	assert.Nil(t, contactByName(t, fixture.jobContacts(t), "Sarah L.").WorkEmail)
}

func TestContactsAreOnlyVisibleForJobsMatchedToTheCandidate(t *testing.T) {
	fixture := newEnrichmentFixture(t)
	require.Equal(t, 2, fixture.enrich(t).Saved)

	intruder := register(t, fixture.router, "intruder@example.com")
	unmatchedJobID := fixture.embeddedJob(t, "Pastry Chef", "We need croissants. "+farJobMarker)
	contactsPath := func(candidateID, jobID string) string {
		return candidatePath(candidateID) + "/jobs/" + jobID + "/contacts"
	}

	otherTenant := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		contactsPath(fixture.candidateID, fixture.jobID),
		nil,
		bearer(intruder.AccessToken),
	)
	ownPathUnmatched := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		contactsPath(*intruder.User.CandidateID, fixture.jobID),
		nil,
		bearer(intruder.AccessToken),
	)
	unmatched := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		contactsPath(fixture.candidateID, unmatchedJobID),
		nil,
		bearer(fixture.userToken),
	)
	malformed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		contactsPath(fixture.candidateID, "not-a-uuid"),
		nil,
		bearer(fixture.userToken),
	)
	anonymous := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		contactsPath(fixture.candidateID, fixture.jobID),
		nil,
		nil,
	)

	assert.Equal(t, http.StatusNotFound, otherTenant.Recorder.Code)
	assert.Equal(
		t,
		http.StatusNotFound,
		ownPathUnmatched.Recorder.Code,
		"a job the candidate is not matched to exposes no contacts",
	)
	assert.Equal(t, http.StatusNotFound, unmatched.Recorder.Code)
	assert.Equal(t, http.StatusNotFound, malformed.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, anonymous.Recorder.Code)
	helpers.AssertMatchesSpec(t, ownPathUnmatched)

	intruderContacts := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(*intruder.User.CandidateID)+"/contacts",
		nil,
		bearer(intruder.AccessToken),
	)
	require.Equal(t, http.StatusOK, intruderContacts.Recorder.Code)
	assert.Empty(t, helpers.DecodeData[[]enrichmentModel.CompanyContact](t, intruderContacts))

	queued := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		contactsPath(fixture.candidateID, fixture.jobID)+"/enrich",
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusAccepted, queued.Recorder.Code, string(queued.Body))
	helpers.AssertMatchesSpec(t, queued)
	assert.Equal(
		t,
		string(workerConfigs.TaskEnrichContacts),
		helpers.DecodeData[enrichmentTypes.RequestEnrichmentResponse](t, queued).TaskName,
	)

	refused := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		contactsPath(*intruder.User.CandidateID, fixture.jobID)+"/enrich",
		nil,
		bearer(intruder.AccessToken),
	)
	assert.Equal(t, http.StatusNotFound, refused.Recorder.Code)
}

func TestMatchesAboveTheThresholdAreHandedToEnrichmentOnceAndOnlyWhenOutreachIsOn(t *testing.T) {
	fixture := newJobFixture(t)
	pinMatchingVectors(setup.Embedder(t))

	nearJobID := fixture.embeddedJob(t, "Senior Go Engineer", "We need Go. "+nearJobMarker)
	fixture.embeddedJob(t, "Pastry Chef", "We need croissants. "+farJobMarker)

	disabled := helpers.Do(
		t, fixture.router, http.MethodPatch, candidatePath(fixture.candidateID),
		map[string]any{"autoOutreachEnabled": false}, bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, disabled.Recorder.Code)

	resume := fixture.processedResume(t, "Alex Candidate "+resumeMarker)
	fannedOut := func() []string {
		rows, err := fixture.deps.Postgres.GetPool().Query(
			context.Background(),
			`SELECT job_id::text FROM job_matches WHERE candidate_id = $1 AND fanned_out_at IS NOT NULL`,
			fixture.candidateID,
		)
		require.NoError(t, err)

		defer rows.Close()

		jobIDs := []string{}

		for rows.Next() {
			var jobID string

			require.NoError(t, rows.Scan(&jobID))
			jobIDs = append(jobIDs, jobID)
		}

		return jobIDs
	}

	_, customErr := fixture.deps.Services.Matching.MatchResume(context.Background(), resume.ID)
	require.Nil(t, customErr)
	assert.Empty(t, fannedOut(), "nothing is handed on while automatic outreach is off")

	enabled := helpers.Do(
		t, fixture.router, http.MethodPatch, candidatePath(fixture.candidateID),
		map[string]any{"autoOutreachEnabled": true}, bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, enabled.Recorder.Code)

	_, customErr = fixture.deps.Services.Matching.MatchResume(context.Background(), resume.ID)
	require.Nil(t, customErr)
	assert.Equal(
		t,
		[]string{nearJobID},
		fannedOut(),
		"only the match above the candidate's threshold is handed on",
	)

	_, customErr = fixture.deps.Services.Matching.MatchResume(context.Background(), resume.ID)
	require.Nil(t, customErr)
	assert.Equal(t, []string{nearJobID}, fannedOut())
}
