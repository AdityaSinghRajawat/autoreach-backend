package unit

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()
	logger := setup.Config(t)

	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(middlewares.Recovery())
	engine.Use(middlewares.RequestLogger(logger))
	engine.Use(middlewares.ErrorHandler())

	return engine
}

func TestEveryErrorCodeHasAnHTTPStatusAndAPublicMessage(t *testing.T) {
	codes := config.GetAllErrCodes()
	require.NotEmpty(t, codes)

	for _, code := range codes {
		assert.True(t, utils.IsErrCodeMapped(code), "%s has no HTTP status mapping", code)
		assert.NotEmpty(
			t,
			config.GetPublicMessageForErrCode(code),
			"%s has no public message",
			code,
		)
	}
}

func TestUnknownErrorCodeFallsBackToInternalServerError(t *testing.T) {
	assert.Equal(
		t,
		http.StatusInternalServerError,
		utils.GetHTTPStatusForErrCode("SOMETHING_UNDECLARED"),
	)
}

func TestCustomErrorBehavesAsAGoError(t *testing.T) {
	sentinel := errors.New("row missing")
	wrapped := fmt.Errorf(
		"find candidate: %w",
		config.NewCustomError(sentinel, config.GetErrCodeNotFound()),
	)

	var customErr *config.CustomError
	require.True(t, errors.As(wrapped, &customErr))
	assert.Equal(t, config.GetErrCodeNotFound(), customErr.ErrCode)
	assert.True(t, errors.Is(wrapped, sentinel))
}

func TestErrorHandlerNeverLeaksInternalDetail(t *testing.T) {
	engine := newEngine(t)
	engine.GET("/boom", func(c *gin.Context) {
		_ = c.Error(errors.New("pq: relation \"candidates\" does not exist"))
	})
	engine.GET("/typed", func(c *gin.Context) {
		utils.HandleCustomError(
			c,
			config.NewCustomError(
				errors.New("select from secret_table failed"),
				config.GetErrCodeConflict(),
			),
		)
	})

	untyped := helpers.Do(t, engine, http.MethodGet, "/boom", nil, nil)
	require.Equal(t, http.StatusInternalServerError, untyped.Recorder.Code)
	assert.NotContains(t, string(untyped.Body), "candidates")
	assert.Equal(
		t,
		string(config.GetErrCodeInternalServerError()),
		helpers.DecodeEnvelope(t, untyped).Error.Code,
	)

	typed := helpers.Do(t, engine, http.MethodGet, "/typed", nil, nil)
	require.Equal(t, http.StatusConflict, typed.Recorder.Code)
	assert.NotContains(t, string(typed.Body), "secret_table")
	assert.Equal(
		t,
		config.GetPublicMessageForErrCode(config.GetErrCodeConflict()),
		helpers.DecodeEnvelope(t, typed).Error.Message,
	)
}

func TestPanicIsRecoveredIntoTheErrorEnvelope(t *testing.T) {
	engine := newEngine(t)
	engine.GET("/panic", func(_ *gin.Context) {
		panic("unexpected nil dependency")
	})

	exchange := helpers.Do(t, engine, http.MethodGet, "/panic", nil, nil)

	require.Equal(t, http.StatusInternalServerError, exchange.Recorder.Code)
	envelope := helpers.DecodeEnvelope(t, exchange)
	assert.False(t, envelope.Success)
	assert.NotContains(t, string(exchange.Body), "nil dependency")
}

func TestInternalSharedSecretRejectsMissingWrongAndUnsetSecrets(t *testing.T) {
	engine := newEngine(t)
	engine.GET(
		"/guarded",
		middlewares.InternalSharedSecret("expected-secret"),
		func(c *gin.Context) {
			utils.BuildAndSendResponse(c, gin.H{"ok": true}, http.StatusOK)
		},
	)
	engine.GET("/unset", middlewares.InternalSharedSecret(""), func(c *gin.Context) {
		utils.BuildAndSendResponse(c, gin.H{"ok": true}, http.StatusOK)
	})

	header := config.GetInternalSecretHeader()

	assert.Equal(
		t,
		http.StatusUnauthorized,
		helpers.Do(t, engine, http.MethodGet, "/guarded", nil, nil).Recorder.Code,
	)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		helpers.Do(
			t,
			engine,
			http.MethodGet,
			"/guarded",
			nil,
			map[string]string{header: "wrong"},
		).Recorder.Code,
	)
	assert.Equal(
		t,
		http.StatusOK,
		helpers.Do(
			t,
			engine,
			http.MethodGet,
			"/guarded",
			nil,
			map[string]string{header: "expected-secret"},
		).Recorder.Code,
	)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		helpers.Do(
			t,
			engine,
			http.MethodGet,
			"/unset",
			nil,
			map[string]string{header: ""},
		).Recorder.Code,
	)
}
