package unit

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

const (
	ownerCandidateID = "6f1c2a3b-0000-4000-8000-000000000001"
	otherCandidateID = "6f1c2a3b-0000-4000-8000-000000000002"
)

func authorizedStatus(
	t *testing.T,
	principal *sharedTypes.Principal,
	permission config.Permission,
	path string,
) int {
	t.Helper()

	engine := newEngine(t)
	handler := func(c *gin.Context) { utils.BuildAndSendResponse(c, gin.H{"ok": true}, http.StatusOK) }
	withPrincipal := func(c *gin.Context) {
		if principal != nil {
			utils.SetPrincipal(c, *principal)
		}

		c.Next()
	}

	engine.GET(
		"/candidates/:"+config.GetCandidateIDParam(),
		withPrincipal,
		middlewares.Authorize(permission),
		handler,
	)
	engine.GET("/global", withPrincipal, middlewares.Authorize(permission), handler)

	return helpers.Do(t, engine, http.MethodGet, path, nil, nil).Recorder.Code
}

func userPrincipal(candidateID string, roles ...string) *sharedTypes.Principal {
	principal := sharedTypes.NewUserPrincipal("user", candidateID, roles)

	return &principal
}

func TestGrantParsingRejectsUnknownPermissionsScopesAndShapes(t *testing.T) {
	setup.Config(t)

	grant, err := config.ParseGrant("resume:read:own")
	require.NoError(t, err)
	assert.Equal(t, config.PermissionResumeRead, grant.Permission)
	assert.Equal(t, config.ScopeOwn, grant.Scope)

	for _, invalid := range []string{
		"", "resume:read", "resume:read:own:extra", "resume:fly:own", "resume:read:everyone", "RESUME:READ:OWN",
	} {
		_, err := config.ParseGrant(invalid)
		assert.ErrorIs(t, err, config.ErrInvalidGrant, "grant %q", invalid)
	}

	for _, permission := range config.GetPermissions() {
		assert.Len(
			t,
			strings.Split(string(permission), ":"),
			2,
			"permissions are resource:action, got %s",
			permission,
		)
	}

	for _, role := range config.GetDefaultSignupRoles() {
		assert.True(t, config.IsKnownRole(role))
	}
}

func TestRolesCombineAndUnknownRolesGrantNothing(t *testing.T) {
	setup.Config(t)

	admin := userPrincipal("", config.RoleAdmin)
	both := userPrincipal(ownerCandidateID, config.RoleCandidate, config.RoleAdmin)
	unknown := userPrincipal(ownerCandidateID, "RETIRED_ROLE")

	for _, permission := range config.GetPermissions() {
		assert.Equal(
			t,
			[]config.AccessScope{config.ScopeAny},
			admin.Grants[permission],
			"admin %s",
			permission,
		)
		assert.Empty(t, unknown.Grants[permission], "a role removed from config grants nothing")
	}

	assert.ElementsMatch(
		t,
		[]config.AccessScope{config.ScopeOwn, config.ScopeAny},
		both.Grants[config.PermissionResumeRead],
		"grants are the union of all roles",
	)
}

func TestOwnScopeReachesOnlyTheOwnersData(t *testing.T) {
	setup.Config(t)

	candidate := userPrincipal(ownerCandidateID, config.RoleCandidate)
	orphan := userPrincipal("", config.RoleCandidate)

	assert.Equal(
		t,
		http.StatusOK,
		authorizedStatus(
			t,
			candidate,
			config.PermissionResumeRead,
			"/candidates/"+ownerCandidateID,
		),
	)
	assert.Equal(
		t,
		http.StatusNotFound,
		authorizedStatus(
			t,
			candidate,
			config.PermissionResumeRead,
			"/candidates/"+otherCandidateID,
		),
		"another tenant's data is hidden, not forbidden",
	)
	assert.Equal(
		t,
		http.StatusNotFound,
		authorizedStatus(t, candidate, config.PermissionResumeRead, "/candidates/not-a-uuid"),
	)
	assert.Equal(
		t,
		http.StatusForbidden,
		authorizedStatus(t, candidate, config.PermissionAPIClientCreate, "/global"),
	)
	assert.Equal(
		t,
		http.StatusForbidden,
		authorizedStatus(t, candidate, config.PermissionResumeRead, "/global"),
		"an own grant never matches a resource without an owner",
	)
	assert.Equal(
		t,
		http.StatusNotFound,
		authorizedStatus(t, orphan, config.PermissionResumeRead, "/candidates/"+ownerCandidateID),
	)
	assert.Equal(
		t,
		http.StatusOK,
		authorizedStatus(t, candidate, config.PermissionJobRead, "/global"),
	)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		authorizedStatus(t, nil, config.PermissionJobRead, "/global"),
	)
}

func TestAnyScopeReachesEveryResourceAndAPIClientsHoldOnlyAllowedGrants(t *testing.T) {
	setup.Config(t)

	admin := userPrincipal("", config.RoleAdmin)
	assert.Equal(
		t,
		http.StatusOK,
		authorizedStatus(
			t,
			admin,
			config.PermissionCandidateUpdate,
			"/candidates/"+otherCandidateID,
		),
	)
	assert.Equal(
		t,
		http.StatusOK,
		authorizedStatus(t, admin, config.PermissionAPIClientCreate, "/global"),
	)

	scraper := sharedTypes.NewAPIClientPrincipal(
		"client-1",
		[]string{"job:ingest:any", "candidate:read:any"},
	)
	assert.Equal(t, config.PrincipalKindAPIClient, scraper.Kind)
	assert.Equal(
		t,
		http.StatusOK,
		authorizedStatus(t, &scraper, config.PermissionJobIngest, "/global"),
	)
	assert.Equal(
		t,
		http.StatusForbidden,
		authorizedStatus(
			t,
			&scraper,
			config.PermissionCandidateRead,
			"/candidates/"+ownerCandidateID,
		),
		"a grant outside the API client allow-list is ignored",
	)
}
