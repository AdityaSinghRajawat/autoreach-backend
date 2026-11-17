package utils

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"syscall"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// NAT64 and 6to4 can translate to a private IPv4 address.
var reservedNetworksI = mustParseNetworks(config.GetReservedNetworks())

func mustParseNetworks(cidrs []string) []*net.IPNet {
	networks := make([]*net.IPNet, 0, len(cidrs))

	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}

		networks = append(networks, network)
	}

	return networks
}

// Refuses private addresses at connect time, which also covers redirects and DNS tricks.
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
		!ip.IsUnspecified() && !slices.ContainsFunc(reservedNetworksI, func(network *net.IPNet) bool {
		return network.Contains(ip)
	})
}
