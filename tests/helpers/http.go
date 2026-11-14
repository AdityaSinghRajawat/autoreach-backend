package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type Envelope struct {
	Success    bool            `json:"success"`
	StatusCode int             `json:"statusCode"`
	Data       json.RawMessage `json:"data"`
	Meta       json.RawMessage `json:"meta"`
	Error      *EnvelopeError  `json:"error"`
}

type EnvelopeError struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Details []config.ErrorDetail `json:"details"`
}

type Exchange struct {
	Request  *http.Request
	Recorder *httptest.ResponseRecorder
	Body     []byte
}

func Do(
	t *testing.T,
	handler http.Handler,
	method, path string,
	body any,
	headers map[string]string,
) Exchange {
	t.Helper()
	payload := encodeBody(t, body)

	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if len(payload) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	request.Body = io.NopCloser(bytes.NewReader(payload))

	return Exchange{Request: request, Recorder: recorder, Body: recorder.Body.Bytes()}
}

func encodeBody(t *testing.T, body any) []byte {
	t.Helper()

	switch typed := body.(type) {
	case nil:
		return nil
	case []byte:
		return typed
	case string:
		return []byte(typed)
	default:
		encoded, err := json.Marshal(body)
		require.NoError(t, err)

		return encoded
	}
}

func DecodeEnvelope(t *testing.T, exchange Exchange) Envelope {
	t.Helper()

	var envelope Envelope
	require.NoError(t, json.Unmarshal(exchange.Body, &envelope), "body: %s", string(exchange.Body))

	return envelope
}

func DecodeData[T any](t *testing.T, exchange Exchange) T {
	t.Helper()

	var data T

	envelope := DecodeEnvelope(t, exchange)
	require.NotEmpty(t, envelope.Data, "envelope has no data: %s", string(exchange.Body))
	require.NoError(t, json.Unmarshal(envelope.Data, &data))

	return data
}
