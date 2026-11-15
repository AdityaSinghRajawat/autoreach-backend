//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

type routeCase struct {
	name   string
	method string
	path   func(candidateID string) string
	body   any
}

func candidateRouteCases() []routeCase {
	resumeID := "00000000-0000-4000-8000-0000000000aa"
	jobID := "00000000-0000-4000-8000-0000000000bb"

	return []routeCase{
		{"get candidate", http.MethodGet, func(id string) string { return candidatePath(id) }, nil},
		{
			"update candidate",
			http.MethodPatch,
			func(id string) string { return candidatePath(id) },
			map[string]any{"fullName": "Changed"},
		},
		{"list resumes", http.MethodGet, func(id string) string { return resumesPath(id) }, nil},
		{
			"get resume",
			http.MethodGet,
			func(id string) string { return resumesPath(id) + "/" + resumeID },
			nil,
		},
		{
			"activate resume",
			http.MethodPatch,
			func(id string) string { return resumesPath(id) + "/" + resumeID },
			map[string]any{"isActive": true},
		},
		{
			"delete resume",
			http.MethodDelete,
			func(id string) string { return resumesPath(id) + "/" + resumeID },
			nil,
		},
		{
			"list matched jobs",
			http.MethodGet,
			func(id string) string { return matchedJobsPath(id, "") },
			nil,
		},
		{
			"list contacts",
			http.MethodGet,
			func(id string) string { return candidatePath(id) + "/contacts" },
			nil,
		},
		{
			"list job contacts",
			http.MethodGet,
			func(id string) string { return candidatePath(id) + "/jobs/" + jobID + "/contacts" },
			nil,
		},
		{
			"list channel accounts",
			http.MethodGet,
			func(id string) string { return channelAccountsPath(id) },
			nil,
		},
		{
			"create channel account",
			http.MethodPost,
			func(id string) string { return channelAccountsPath(id) },
			map[string]any{"emailAddress": "invalid"},
		},
		{
			"delete channel account",
			http.MethodDelete,
			func(id string) string { return channelAccountsPath(id) + "/" + resumeID },
			nil,
		},
		{
			"verify channel account",
			http.MethodPost,
			func(id string) string { return channelAccountsPath(id) + "/" + resumeID + "/verify" },
			nil,
		},
		{"list templates", http.MethodGet, templatesPath, nil},
		{"create template", http.MethodPost, templatesPath, map[string]any{"name": "x"}},
		{
			"get template",
			http.MethodGet,
			func(id string) string { return templatesPath(id) + "/" + resumeID },
			nil,
		},
		{
			"update template",
			http.MethodPatch,
			func(id string) string { return templatesPath(id) + "/" + resumeID },
			map[string]any{"name": "x"},
		},
		{
			"delete template",
			http.MethodDelete,
			func(id string) string { return templatesPath(id) + "/" + resumeID },
			nil,
		},
		{"list campaigns", http.MethodGet, campaignsPath, nil},
		{"create campaign", http.MethodPost, campaignsPath, map[string]any{"name": "x"}},
		{
			"get campaign",
			http.MethodGet,
			func(id string) string { return campaignsPath(id) + "/" + resumeID },
			nil,
		},
		{
			"update campaign",
			http.MethodPatch,
			func(id string) string { return campaignsPath(id) + "/" + resumeID },
			map[string]any{"name": "x"},
		},
		{
			"list outreach messages",
			http.MethodGet,
			func(id string) string { return candidatePath(id) + "/outreach-messages" },
			nil,
		},
		{
			"create outreach messages",
			http.MethodPost,
			func(id string) string { return candidatePath(id) + "/outreach-messages" },
			map[string]any{"jobId": "x"},
		},
		{
			"get outreach message",
			http.MethodGet,
			func(id string) string { return candidatePath(id) + "/outreach-messages/" + resumeID },
			nil,
		},
		{
			"cancel outreach message",
			http.MethodPatch,
			func(id string) string { return candidatePath(id) + "/outreach-messages/" + resumeID },
			map[string]any{"status": "CANCELLED"},
		},
		{
			"confirm outreach message",
			http.MethodPost,
			func(id string) string { return candidatePath(id) + "/outreach-messages/" + resumeID + "/confirm" },
			map[string]any{"status": "SENT"},
		},
		{
			"request enrichment",
			http.MethodPost,
			func(id string) string { return candidatePath(id) + "/jobs/" + jobID + "/contacts/enrich" },
			nil,
		},
	}
}

func TestEveryCandidateRouteEnforcesOwnershipForEveryPrincipal(t *testing.T) {
	fixture := newJobFixture(t)
	other := register(t, fixture.router, "other@example.com")

	adminToken, _, err := utils.IssueAccessToken(adminPrincipal(), utils.GetCurrTimeStamp())
	require.NoError(t, err)

	retiredRoleToken, _, err := utils.IssueAccessToken(
		sharedTypes.NewUserPrincipal(
			other.User.UserID,
			*other.User.CandidateID,
			[]string{"RETIRED_ROLE"},
		),
		utils.GetCurrTimeStamp(),
	)
	require.NoError(t, err)

	for _, route := range candidateRouteCases() {
		t.Run(route.name, func(t *testing.T) {
			do := func(candidateID string, headers map[string]string) int {
				exchange := helpers.Do(
					t,
					fixture.router,
					route.method,
					route.path(candidateID),
					route.body,
					headers,
				)

				return exchange.Recorder.Code
			}

			assert.Equal(t, http.StatusUnauthorized, do(fixture.candidateID, nil), "anonymous")
			assert.Equal(
				t,
				http.StatusNotFound,
				do(fixture.candidateID, bearer(other.AccessToken)),
				"another candidate",
			)
			assert.Equal(
				t,
				http.StatusNotFound,
				do("not-a-uuid", bearer(fixture.userToken)),
				"malformed candidate id",
			)
			assert.Equal(
				t,
				http.StatusForbidden,
				do(fixture.candidateID, bearer(retiredRoleToken)),
				"role without the permission",
			)
			assert.NotContains(
				t,
				[]int{http.StatusUnauthorized, http.StatusForbidden},
				do(fixture.candidateID, bearer(fixture.userToken)),
				"the owner passes authorisation",
			)
			assert.NotContains(
				t,
				[]int{http.StatusUnauthorized, http.StatusForbidden},
				do(fixture.candidateID, bearer(adminToken)),
				"an admin passes authorisation through the any scope",
			)
		})
	}
}

func TestRolesAreStoredPerUserAndCarriedInTheSession(t *testing.T) {
	fixture := newJobFixture(t)

	login := helpers.Do(t, fixture.router, http.MethodPost, loginPath, map[string]any{
		"email": adminEmail, "password": adminPassword,
	}, nil)
	require.Equal(t, http.StatusOK, login.Recorder.Code, string(login.Body))
	helpers.AssertMatchesSpec(t, login)

	me := helpers.Do(t, fixture.router, http.MethodGet, mePath, nil, bearer(
		helpers.DecodeData[struct {
			AccessToken string `json:"accessToken"`
		}](t, login).AccessToken,
	))
	require.Equal(t, http.StatusOK, me.Recorder.Code)
	helpers.AssertMatchesSpec(t, me)
	assert.JSONEq(t, `["ADMIN"]`, extractField(t, me, "roles"))

	var rows int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		t.Context(), `SELECT COUNT(*) FROM users WHERE roles = ARRAY[$1]`, config.RoleCandidate,
	).Scan(&rows))
	assert.Equal(t, 1, rows, "the registered candidate holds exactly the candidate role")
}

func TestApiClientPrincipalCannotReachUserRoutes(t *testing.T) {
	fixture := newJobFixture(t)

	viaAPIKey := helpers.Do(
		t, fixture.router, http.MethodGet, candidatePath(fixture.candidateID), nil,
		apiKeyHeader(fixture.apiKey),
	)
	adminViaAPIKey := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		apiClientsPath,
		nil,
		apiKeyHeader(fixture.apiKey),
	)

	assert.Equal(
		t,
		http.StatusUnauthorized,
		viaAPIKey.Recorder.Code,
		"user routes only accept user tokens",
	)
	assert.Equal(t, http.StatusUnauthorized, adminViaAPIKey.Recorder.Code)
}

func extractField(t *testing.T, exchange helpers.Exchange, field string) string {
	t.Helper()

	data := helpers.DecodeData[map[string]any](t, exchange)
	value, found := data[field]
	require.True(t, found, "field %s missing", field)

	encoded, err := jsonMarshal(value)
	require.NoError(t, err)

	return encoded
}

func jsonMarshal(value any) (string, error) {
	encoded, err := json.Marshal(value)

	return string(encoded), err
}
