//go:build integration

package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

// Routes reached without a user token: login and signup, the ingest API (API key), webhooks
// (signatures) and the per-message tracking and unsubscribe links (signed URLs).
var publicRoutePrefixesI = []string{
	"/api/v1/auth/",
	"/api/v1/jobs/ingest",
	"/api/v1/webhooks/",
}

type registeredRoute struct {
	method string
	path   string
}

func apiRoutes(t *testing.T, router http.Handler) []registeredRoute {
	t.Helper()

	engine, isEngine := router.(*gin.Engine)
	require.True(t, isEngine)

	routes := []registeredRoute{}

	for _, route := range engine.Routes() {
		if !strings.HasPrefix(route.Path, config.GetAPIBasePath()) {
			continue
		}

		routes = append(routes, registeredRoute{method: route.Method, path: route.Path})
	}

	require.NotEmpty(t, routes)

	return routes
}

func isPublicRoute(path string) bool {
	for _, prefix := range publicRoutePrefixesI {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// fillPath puts the candidate into :candidateId and random IDs into every other parameter.
func fillPath(path, candidateID string) string {
	segments := strings.Split(path, "/")

	for index, segment := range segments {
		name, isParam := strings.CutPrefix(segment, ":")
		if !isParam {
			continue
		}

		segments[index] = uuid.NewString()
		if name == config.GetCandidateIDParam() {
			segments[index] = candidateID
		}
	}

	return strings.Join(segments, "/")
}

func TestEveryCandidateRouteHidesOtherCandidatesData(t *testing.T) {
	fixture := newJobFixture(t)
	intruder := register(t, fixture.router, "intruder@example.com")
	checked := 0

	for _, route := range apiRoutes(t, fixture.router) {
		target := ""

		switch {
		case strings.Contains(route.path, ":"+config.GetCandidateIDParam()):
			target = fillPath(route.path, fixture.candidateID)
		case strings.HasPrefix(route.path, config.GetAPIBasePath()+"/dashboard"):
			target = route.path + "?" + config.GetCandidateIDQueryParam() + "=" + fixture.candidateID
		default:
			continue
		}

		exchange := helpers.Do(
			t,
			fixture.router,
			route.method,
			target,
			map[string]any{},
			bearer(intruder.AccessToken),
		)
		assert.Equal(
			t,
			http.StatusNotFound,
			exchange.Recorder.Code,
			"%s %s must not reveal another candidate's data: %s",
			route.method,
			route.path,
			string(exchange.Body),
		)

		checked++
	}

	assert.Greater(t, checked, 30, "the matrix should cover every candidate-scoped route")
}

func TestEveryNonPublicRouteRequiresAuthentication(t *testing.T) {
	fixture := newJobFixture(t)

	for _, route := range apiRoutes(t, fixture.router) {
		if isPublicRoute(route.path) {
			continue
		}

		exchange := helpers.Do(
			t,
			fixture.router,
			route.method,
			fillPath(route.path, fixture.candidateID),
			map[string]any{},
			nil,
		)
		assert.Equal(
			t,
			http.StatusUnauthorized,
			exchange.Recorder.Code,
			"%s %s",
			route.method,
			route.path,
		)
	}
}

func TestEveryAdminRouteRejectsCandidates(t *testing.T) {
	fixture := newJobFixture(t)
	checked := 0

	for _, route := range apiRoutes(t, fixture.router) {
		if !strings.HasPrefix(route.path, config.GetAPIBasePath()+"/admin") {
			continue
		}

		exchange := helpers.Do(
			t,
			fixture.router,
			route.method,
			fillPath(route.path, fixture.candidateID),
			map[string]any{},
			bearer(fixture.userToken),
		)
		assert.Equal(
			t,
			http.StatusForbidden,
			exchange.Recorder.Code,
			"%s %s",
			route.method,
			route.path,
		)

		checked++
	}

	assert.Greater(t, checked, 10)
}
