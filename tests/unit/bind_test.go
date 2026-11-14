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
)

type bindItem struct {
	CompanyName string `json:"companyName" validate:"required,max=10"`
}

type bindPayload struct {
	Email string     `json:"email" validate:"required,email"`
	Items []bindItem `json:"items" validate:"required,min=1,dive"`
}

func newBindEngine(t *testing.T, maxBodyBytes int64) *gin.Engine {
	t.Helper()
	engine := newEngine(t)
	engine.POST("/bind", middlewares.BodyLimit(maxBodyBytes), func(c *gin.Context) {
		var payload bindPayload
		if customErr := utils.BindAndValidate(c, &payload); customErr != nil {
			utils.HandleCustomError(c, customErr)

			return
		}

		utils.BuildAndSendResponse(c, payload, http.StatusCreated)
	})

	return engine
}

func TestBindAcceptsAValidPayload(t *testing.T) {
	exchange := helpers.Do(
		t,
		newBindEngine(t, 1024),
		http.MethodPost,
		"/bind",
		`{"email":"a@b.co","items":[{"companyName":"Acme"}]}`,
		nil,
	)

	require.Equal(t, http.StatusCreated, exchange.Recorder.Code)
	assert.Equal(t, "a@b.co", helpers.DecodeData[bindPayload](t, exchange).Email)
}

func TestBindReturnsBadRequestForMalformedJSON(t *testing.T) {
	engine := newBindEngine(t, 1024)

	for _, body := range []string{`{"email":`, `not json`, `{"email":"a@b.co","items":[{"companyName":"Acme"}]} trailing`, ``} {
		exchange := helpers.Do(t, engine, http.MethodPost, "/bind", body, nil)
		require.Equal(t, http.StatusBadRequest, exchange.Recorder.Code, "body %q", body)
		assert.Equal(
			t,
			string(config.GetErrCodeBadRequest()),
			helpers.DecodeEnvelope(t, exchange).Error.Code,
		)
	}
}

func TestBindReturnsUnprocessableEntityWithCamelCaseFieldPaths(t *testing.T) {
	exchange := helpers.Do(
		t,
		newBindEngine(t, 1024),
		http.MethodPost,
		"/bind",
		`{"email":"nope","items":[{"companyName":""}]}`,
		nil,
	)

	require.Equal(t, http.StatusUnprocessableEntity, exchange.Recorder.Code)
	envelope := helpers.DecodeEnvelope(t, exchange)
	require.NotNil(t, envelope.Error)
	assert.Equal(t, string(config.GetErrCodeValidation()), envelope.Error.Code)
	assert.ElementsMatch(t, []config.ErrorDetail{
		{Field: "email", Issue: "must be a valid email address"},
		{Field: "items[0].companyName", Issue: "is required"},
	}, envelope.Error.Details)
}

func TestBindReturnsPayloadTooLargeWhenTheBodyExceedsTheLimit(t *testing.T) {
	body := `{"email":"a@b.co","items":[{"companyName":"` + strings.Repeat("x", 200) + `"}]}`

	exchange := helpers.Do(t, newBindEngine(t, 64), http.MethodPost, "/bind", body, nil)

	require.Equal(t, http.StatusRequestEntityTooLarge, exchange.Recorder.Code)
	assert.Equal(
		t,
		string(config.GetErrCodePayloadTooLarge()),
		helpers.DecodeEnvelope(t, exchange).Error.Code,
	)
}

func TestPageParamsNormalizeAndMeta(t *testing.T) {
	newEngine(t)

	defaults := sharedTypes.PageParams{}.Normalize()
	assert.Equal(t, config.GetDefaultPage(), defaults.Page)
	assert.Equal(t, config.GetDefaultPerPage(), defaults.PerPage)

	clamped := sharedTypes.PageParams{Page: 3, PerPage: 10_000}.Normalize()
	assert.Equal(t, config.GetMaxPerPage(), clamped.PerPage)
	assert.Equal(t, 2*config.GetMaxPerPage(), clamped.Offset())

	meta := sharedTypes.NewPaginationMeta(sharedTypes.PageParams{Page: 1, PerPage: 10}, 128)
	assert.Equal(
		t,
		sharedTypes.PaginationMeta{Page: 1, PerPage: 10, TotalItems: 128, TotalPages: 13},
		meta,
	)
	assert.Equal(
		t,
		0,
		sharedTypes.NewPaginationMeta(sharedTypes.PageParams{Page: 1, PerPage: 10}, 0).TotalPages,
	)
}
