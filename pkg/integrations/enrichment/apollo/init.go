package apollo

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	searchRequest struct {
		Domains []string `json:"q_organization_domains_list"`
		Titles  []string `json:"person_titles"`
		Page    int      `json:"page"`
		PerPage int      `json:"per_page"`
	}

	SearchResponse struct {
		People []Person `json:"people"`
	}

	matchRequest struct {
		ID                   string `json:"id"`
		RevealPersonalEmails bool   `json:"reveal_personal_emails"`
	}

	MatchResponse struct {
		Person *Person `json:"person"`
	}

	Person struct {
		ID           string        `json:"id"`
		FirstName    string        `json:"first_name"`
		LastName     string        `json:"last_name"`
		Name         string        `json:"name"`
		Title        string        `json:"title"`
		Email        string        `json:"email"`
		EmailStatus  string        `json:"email_status"`
		LinkedinURL  string        `json:"linkedin_url"`
		Organization *Organization `json:"organization"`
	}

	organizationSearchRequest struct {
		Name    string `json:"q_organization_name"`
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
	}

	OrganizationSearchResponse struct {
		Organizations []Organization `json:"organizations"`
		Accounts      []Organization `json:"accounts"`
	}

	Organization struct {
		Name          string `json:"name"`
		PrimaryDomain string `json:"primary_domain"`
	}
)

type ApolloIntegration struct {
	httpClient *utils.HTTPClient
	apiKey     string
}

func NewApolloIntegration(policy config.ProviderPolicy) *ApolloIntegration {
	return &ApolloIntegration{
		httpClient: utils.NewHTTPClient(policy.BaseURL, policy.Timeout),
		apiKey:     policy.APIKey,
	}
}
