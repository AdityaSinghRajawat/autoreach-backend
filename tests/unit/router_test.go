package unit

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/mocks"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func newRouter(t *testing.T) http.Handler {
	t.Helper()
	logger := setup.Config(t)

	return routes.NewRoutes(&dependencies.Dependencies{
		Logger:      logger,
		RateLimiter: mocks.NewAllowingLimiter(),
	})
}

func TestHealthReturnsTheStandardEnvelope(t *testing.T) {
	exchange := helpers.Do(t, newRouter(t), http.MethodGet, "/health", nil, nil)

	require.Equal(t, http.StatusOK, exchange.Recorder.Code)
	helpers.AssertMatchesSpec(t, exchange)

	envelope := helpers.DecodeEnvelope(t, exchange)
	assert.True(t, envelope.Success)
	assert.Equal(t, http.StatusOK, envelope.StatusCode)
	assert.Nil(t, envelope.Error)
	assert.Equal(t, "ok", helpers.DecodeData[sharedTypes.HealthResponse](t, exchange).Status)
}

func TestEveryResponseCarriesACorrelationID(t *testing.T) {
	router := newRouter(t)

	generated := helpers.Do(t, router, http.MethodGet, "/health", nil, nil)
	assert.NotEmpty(t, generated.Recorder.Header().Get(config.GetCorrelationIDHeader()))

	supplied := "7b0e7a3e-52f5-4f0b-9a55-0e2f6a3a9d11"
	echoed := helpers.Do(
		t,
		router,
		http.MethodGet,
		"/health",
		nil,
		map[string]string{config.GetCorrelationIDHeader(): supplied},
	)
	assert.Equal(t, supplied, echoed.Recorder.Header().Get(config.GetCorrelationIDHeader()))

	replaced := helpers.Do(
		t,
		router,
		http.MethodGet,
		"/health",
		nil,
		map[string]string{config.GetCorrelationIDHeader(): "not-a-uuid\nforged"},
	)
	assert.NotEqual(
		t,
		"not-a-uuid\nforged",
		replaced.Recorder.Header().Get(config.GetCorrelationIDHeader()),
	)
}

func TestUnknownRouteReturnsTheNotFoundEnvelope(t *testing.T) {
	exchange := helpers.Do(t, newRouter(t), http.MethodGet, "/api/v1/does-not-exist", nil, nil)

	require.Equal(t, http.StatusNotFound, exchange.Recorder.Code)
	envelope := helpers.DecodeEnvelope(t, exchange)
	assert.False(t, envelope.Success)
	assert.Equal(t, http.StatusNotFound, envelope.StatusCode)
	require.NotNil(t, envelope.Error)
	assert.Equal(t, string(config.GetErrCodeNotFound()), envelope.Error.Code)
}

func TestWrongMethodReturnsTheMethodNotAllowedEnvelope(t *testing.T) {
	exchange := helpers.Do(t, newRouter(t), http.MethodDelete, "/health", nil, nil)

	require.Equal(t, http.StatusMethodNotAllowed, exchange.Recorder.Code)
	envelope := helpers.DecodeEnvelope(t, exchange)
	require.NotNil(t, envelope.Error)
	assert.Equal(t, string(config.GetErrCodeMethodNotAllowed()), envelope.Error.Code)
}

func TestOpenAPISpecIsServed(t *testing.T) {
	exchange := helpers.Do(t, newRouter(t), http.MethodGet, "/openapi.json", nil, nil)

	require.Equal(t, http.StatusOK, exchange.Recorder.Code)
	assert.Contains(t, string(exchange.Body), `"openapi"`)
}

var ginParameterPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

func TestEveryAPIRouteIsDocumentedInTheSpec(t *testing.T) {
	engine, isEngine := newRouter(t).(*gin.Engine)
	require.True(t, isEngine)
	documented := helpers.DocumentedOperations(t)

	for _, route := range engine.Routes() {
		if route.Path != "/health" && !strings.HasPrefix(route.Path, config.GetAPIBasePath()) {
			continue
		}

		specPath := ginParameterPattern.ReplaceAllString(route.Path, "{$1}")
		_, found := documented[route.Method+" "+specPath]
		assert.True(
			t,
			found,
			"%s %s is registered but missing from openapi.json",
			route.Method,
			specPath,
		)
	}
}
