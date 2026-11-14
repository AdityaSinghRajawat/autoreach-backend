package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

const maxErrorBodyBytes = 2048

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("upstream responded with status %d", e.StatusCode)
}

func (e *APIError) IsTransient() bool {
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode >= http.StatusInternalServerError
}

type HttpClient struct {
	client  *http.Client
	baseURL string
}

func NewHttpClient(baseURL string, timeout time.Duration) *HttpClient {
	return &HttpClient{
		client:  &http.Client{Timeout: timeout},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *HttpClient) Get(
	ctx context.Context,
	endpoint string,
	headers map[string]string,
	result any,
) error {
	return c.do(ctx, http.MethodGet, endpoint, headers, nil, result)
}

func (c *HttpClient) Post(
	ctx context.Context,
	endpoint string,
	headers map[string]string,
	payload, result any,
) error {
	return c.do(ctx, http.MethodPost, endpoint, headers, payload, result)
}

func (c *HttpClient) do(
	ctx context.Context,
	method, endpoint string,
	headers map[string]string,
	payload, result any,
) error {
	var body io.Reader

	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}

		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	request.Header.Set("Accept", "application/json")

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		excerpt, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))

		return &APIError{StatusCode: response.StatusCode, Body: string(excerpt)}
	}

	if result == nil {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

func ClassifyUpstreamError(err error) config.CustomErrorCode {
	var apiErr *APIError
	if errors.As(err, &apiErr) && !apiErr.IsTransient() {
		return config.GetErrCodeUpstreamRejected()
	}

	if errors.Is(err, config.ErrMissingCredential) {
		return config.GetErrCodeUpstreamRejected()
	}

	return config.GetErrCodeServiceUnavailable()
}
