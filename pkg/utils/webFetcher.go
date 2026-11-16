package utils

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

var carrierGradeNATI = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// WebFetcher reads public web pages. Its default transport refuses to connect to loopback, private
// and link-local addresses, which also covers redirects and DNS tricks.
type WebFetcher struct {
	client    *http.Client
	userAgent string
	maxBytes  int64
}

func NewWebFetcher(
	timeout time.Duration,
	maxBytes int64,
	userAgent string,
	transport http.RoundTripper,
) *WebFetcher {
	if transport == nil {
		transport = publicOnlyTransport(timeout)
	}

	return &WebFetcher{
		client:    &http.Client{Timeout: timeout, Transport: transport},
		userAgent: userAgent,
		maxBytes:  maxBytes,
	}
}

func (f *WebFetcher) UserAgent() string { return f.userAgent }

func (f *WebFetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	request.Header.Set(config.GetUserAgentHeader(), f.userAgent)

	response, err := f.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{StatusCode: response.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, f.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rawURL, err)
	}

	if int64(len(body)) > f.maxBytes {
		return nil, fmt.Errorf("read %s: %w", rawURL, config.ErrResponseTooLarge)
	}

	return body, nil
}

func publicOnlyTransport(timeout time.Duration) *http.Transport {
	dialer := &net.Dialer{Timeout: timeout, Control: publicAddressOnly}

	return &http.Transport{DialContext: dialer.DialContext}
}

// publicAddressOnly runs after DNS resolution, so a hostname that resolves to a private address, or a
// redirect to one, is refused too.
func publicAddressOnly(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip == nil || !isPublicIP(ip) {
		return config.ErrPrivateNetworkAddress
	}

	return nil
}

func isPublicIP(ip net.IP) bool {
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() && !ip.IsInterfaceLocalMulticast() && !ip.IsMulticast() &&
		!ip.IsUnspecified() && !carrierGradeNATI.Contains(ip)
}
