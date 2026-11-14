package helpers

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/require"

	swaggerDocs "github.com/AdityaSinghRajawat/autoreach-backend/docs/swagger"
)

var (
	specOnce   sync.Once
	specErr    error
	specDoc    *openapi3.T
	specRouter routers.Router
)

func OpenAPISpec(t *testing.T) *openapi3.T {
	t.Helper()
	loadSpec()
	require.NoError(t, specErr)

	return specDoc
}

func loadSpec() {
	specOnce.Do(func() {
		loader := openapi3.NewLoader()

		specDoc, specErr = loader.LoadFromData(swaggerDocs.GetOpenAPISpec())
		if specErr != nil {
			return
		}

		if specErr = specDoc.Validate(loader.Context); specErr != nil {
			return
		}

		specDoc.Servers = nil
		specRouter, specErr = gorillamux.NewRouter(specDoc)
	})
}

func AssertMatchesSpec(t *testing.T, exchange Exchange) {
	t.Helper()
	loadSpec()
	require.NoError(t, specErr)

	route, pathParams, err := specRouter.FindRoute(exchange.Request)
	require.NoError(
		t,
		err,
		"route %s %s is not documented in openapi.json",
		exchange.Request.Method,
		exchange.Request.URL.Path,
	)

	requestInput := &openapi3filter.RequestValidationInput{
		Request:    exchange.Request,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	responseInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestInput,
		Status:                 exchange.Recorder.Code,
		Header:                 exchange.Recorder.Header(),
		Body:                   io.NopCloser(bytes.NewReader(exchange.Body)),
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
		},
	}
	require.NoError(
		t,
		openapi3filter.ValidateResponse(context.Background(), responseInput),
		"response does not match openapi.json: %s %s -> %d %s",
		exchange.Request.Method,
		exchange.Request.URL.Path,
		exchange.Recorder.Code,
		string(exchange.Body),
	)
}

func DocumentedOperations(t *testing.T) map[string]struct{} {
	t.Helper()

	documented := map[string]struct{}{}

	for path, item := range OpenAPISpec(t).Paths.Map() {
		for method := range item.Operations() {
			documented[method+" "+path] = struct{}{}
		}
	}

	return documented
}
