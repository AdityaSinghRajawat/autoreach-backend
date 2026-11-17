package helpers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type EnrichmentStub struct {
	Server *httptest.Server

	mutex              sync.Mutex
	SearchStatus       int
	MatchStatus        int
	ContactOutStatus   int
	SignalHireStatus   int
	SignalHireSearch   int
	People             []map[string]any
	Matches            map[string]map[string]any
	Profiles           map[string]map[string]any
	Organizations      map[string][]map[string]any
	SignalHireProfiles []map[string]any
	calls              map[string]int
	signalHireItems    []string
	signalHireRequests int
	signalHireQueries  []map[string]any
}

func NewEnrichmentStub(t *testing.T) *EnrichmentStub {
	t.Helper()

	stub := &EnrichmentStub{
		Matches:       map[string]map[string]any{},
		Profiles:      map[string]map[string]any{},
		Organizations: map[string][]map[string]any{},
		calls:         map[string]int{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /apollo/mixed_people/api_search", stub.handleSearch)
	mux.HandleFunc("POST /apollo/people/match", stub.handleMatch)
	mux.HandleFunc("POST /apollo/mixed_companies/search", stub.handleOrganizations)
	mux.HandleFunc("GET /contactout/linkedin/enrich", stub.handleContactOut)
	mux.HandleFunc("POST /signalhire/candidate/search", stub.handleSignalHire)
	mux.HandleFunc("POST /signalhire/candidate/searchByQuery", stub.handleSignalHireSearch)

	stub.Server = httptest.NewServer(mux)
	t.Cleanup(stub.Server.Close)

	return stub
}

func (s *EnrichmentStub) Policies() []config.ProviderPolicy {
	policy := func(code, path string) config.ProviderPolicy {
		return config.ProviderPolicy{
			Code:         code,
			APIKey:       "test-" + strings.ToLower(code),
			BaseURL:      s.Server.URL + path,
			Timeout:      5 * time.Second,
			RateLimit:    1000,
			RateWindow:   time.Minute,
			DailyCredits: 1000,
		}
	}

	return []config.ProviderPolicy{
		policy(config.ProviderApollo, "/apollo"),
		policy(config.ProviderSignalHire, "/signalhire"),
		policy(config.ProviderContactOut, "/contactout"),
	}
}

func (s *EnrichmentStub) Calls(name string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.calls[name]
}

func (s *EnrichmentStub) SignalHireItems() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return append([]string(nil), s.signalHireItems...)
}

func (s *EnrichmentStub) SignalHireQueries() []map[string]any {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return append([]map[string]any(nil), s.signalHireQueries...)
}

func (s *EnrichmentStub) Set(change func(stub *EnrichmentStub)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	change(s)
}

func (s *EnrichmentStub) record(name string, status int) int {
	s.calls[name]++

	if status == 0 {
		return http.StatusOK
	}

	return status
}

func (s *EnrichmentStub) handleSearch(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	status := s.record("apollo.search", s.SearchStatus)
	if request.Header.Get("X-Api-Key") == "" {
		status = http.StatusUnauthorized
	}

	writeJSON(writer, status, map[string]any{"people": s.People})
}

func (s *EnrichmentStub) handleOrganizations(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	body := struct {
		Name string `json:"q_organization_name"`
	}{}
	_ = json.NewDecoder(request.Body).Decode(&body)

	status := s.record("apollo.organizations", 0)
	organizations := s.Organizations[body.Name]

	if organizations == nil {
		organizations = []map[string]any{}
	}

	writeJSON(writer, status, map[string]any{"organizations": organizations, "accounts": []any{}})
}

func (s *EnrichmentStub) handleMatch(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	body := struct {
		ID string `json:"id"`
	}{}
	_ = json.NewDecoder(request.Body).Decode(&body)

	status := s.record("apollo.match", s.MatchStatus)
	writeJSON(writer, status, map[string]any{"person": s.Matches[body.ID]})
}

func (s *EnrichmentStub) handleContactOut(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	status := s.record("contactout.enrich", s.ContactOutStatus)

	profile, found := s.Profiles[request.URL.Query().Get("profile")]
	if status == http.StatusOK && !found {
		status = http.StatusNotFound
	}

	writeJSON(writer, status, map[string]any{"status_code": status, "profile": profile})
}

func (s *EnrichmentStub) handleSignalHire(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	body := struct {
		Items       []string `json:"items"`
		CallbackURL string   `json:"callbackUrl"`
	}{}
	_ = json.NewDecoder(request.Body).Decode(&body)

	status := s.record("signalhire.request", s.SignalHireStatus)
	if status == http.StatusOK {
		status = http.StatusCreated
	}

	s.signalHireItems = append(s.signalHireItems, body.Items...)
	s.signalHireRequests++

	writeJSON(writer, status, map[string]any{"requestId": s.signalHireRequests})
}

func (s *EnrichmentStub) handleSignalHireSearch(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	body := map[string]any{}
	_ = json.NewDecoder(request.Body).Decode(&body)

	status := s.record("signalhire.search", s.SignalHireSearch)
	if request.Header.Get("apikey") == "" {
		status = http.StatusUnauthorized
	}

	s.signalHireQueries = append(s.signalHireQueries, body)

	profiles := s.SignalHireProfiles
	if profiles == nil {
		profiles = []map[string]any{}
	}

	writeJSON(
		writer,
		status,
		map[string]any{"requestId": 1, "total": len(profiles), "profiles": profiles},
	)
}

func SignalHireRequestID(sequence int) string {
	return strconv.Itoa(sequence)
}

func writeJSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

type FakeMXResolver struct {
	Domains map[string]bool
}

func (r FakeMXResolver) LookupMX(_ context.Context, domain string) ([]*net.MX, error) {
	if r.Domains[domain] {
		return []*net.MX{{Host: "mx." + domain, Pref: 10}}, nil
	}

	return nil, &net.DNSError{Err: "no such host", Name: domain, IsNotFound: true}
}
