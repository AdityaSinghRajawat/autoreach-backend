package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

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

func (c *HttpClient) PostMultipart(
	ctx context.Context,
	endpoint string,
	headers map[string]string,
	fields map[string][]string,
	files []sharedTypes.MultipartFile,
	result any,
) error {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	for name, values := range fields {
		for _, value := range values {
			if err := writer.WriteField(name, value); err != nil {
				return fmt.Errorf("encode multipart field: %w", err)
			}
		}
	}

	for _, file := range files {
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Disposition": {
				`form-data; name="` + file.FieldName + `"; filename="` + file.FileName + `"`,
			},
			"Content-Type": {file.ContentType},
		})
		if err != nil {
			return fmt.Errorf("encode multipart file: %w", err)
		}

		if _, err := part.Write(file.Content); err != nil {
			return fmt.Errorf("encode multipart file: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("encode multipart body: %w", err)
	}

	return c.send(
		ctx,
		http.MethodPost,
		endpoint,
		headers,
		&body,
		writer.FormDataContentType(),
		result,
	)
}

func (c *HttpClient) do(
	ctx context.Context,
	method, endpoint string,
	headers map[string]string,
	payload, result any,
) error {
	if payload == nil {
		return c.send(ctx, method, endpoint, headers, nil, "", result)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request body: %w", err)
	}

	return c.send(
		ctx,
		method,
		endpoint,
		headers,
		bytes.NewReader(encoded),
		config.GetJSONContentType(),
		result,
	)
}

func (c *HttpClient) send(
	ctx context.Context,
	method, endpoint string,
	headers map[string]string,
	body io.Reader,
	contentType string,
	result any,
) error {
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	if contentType != "" {
		request.Header.Set(config.GetContentTypeHeader(), contentType)
	}

	request.Header.Set(config.GetAcceptHeader(), config.GetJSONContentType())

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", withoutQuery(err))
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		excerpt, _ := io.ReadAll(io.LimitReader(response.Body, config.GetMaxErrorBodyBytes()))

		return &APIError{StatusCode: response.StatusCode, Body: string(excerpt)}
	}

	if result == nil {
		return nil
	}

	limited := io.LimitReader(response.Body, config.GetMaxResponseBodyBytes()+1)
	if err := json.NewDecoder(limited).Decode(result); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

// withoutQuery drops the query string from a transport error, because some APIs take their key as a
// query parameter and the error would otherwise carry it into logs and task records.
func withoutQuery(err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err
	}

	redacted := *urlErr
	if parsed, parseErr := url.Parse(urlErr.URL); parseErr == nil {
		parsed.RawQuery = ""
		redacted.URL = parsed.String()
	}

	return &redacted
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
