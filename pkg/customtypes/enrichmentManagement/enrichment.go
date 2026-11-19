package enrichmentManagement

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type DiscoveryQuery struct {
	CompanyDomain string
	CompanyName   string
	Titles        []string
	Limit         int
}

type Person struct {
	Source           string
	ProviderPersonID string
	FullName         string
	Title            string
	LinkedinURL      string
	CompanyName      string
	Email            string
	EmailVerified    bool
}

type SelectedPerson struct {
	Person     Person
	Role       string
	SourceCode string
}

type ContactDetails struct {
	FullName      string
	Title         string
	Email         string
	EmailVerified bool
	Phone         string
	LinkedinURL   string
	Raw           json.RawMessage
}

type CallbackResult struct {
	Item    string
	Found   bool
	Details ContactDetails
}

type ListContactsParams struct {
	sharedTypes.PageParams
}

type EnrichJobContactsResult struct {
	JobID        string `json:"jobId"`
	Reused       int    `json:"reused"`
	Discovered   int    `json:"discovered"`
	Saved        int    `json:"saved"`
	PendingAsync int    `json:"pendingAsync"`
}

type RequestEnrichmentResponse struct {
	JobID    string `json:"jobId"`
	TaskName string `json:"taskName"`
	Status   string `json:"status"`
}

type WebhookAckResponse struct {
	Received bool `json:"received"`
}

func DailyCreditsKey(providerCode string) string {
	return utils.BuildKey(
		config.GetKeyDelimiter(),
		config.GetKeyPrefix(),
		config.GetEnrichmentCreditNamespace(),
		providerCode,
		utils.GetCurrTimeStamp().Format(time.DateOnly),
	)
}

func (p Person) HasLinkedin() bool {
	return p.LinkedinURL != ""
}

func (p Person) HasEmail() bool {
	return p.Email != ""
}

func (p Person) KnownDetails() *ContactDetails {
	return &ContactDetails{
		FullName:      p.FullName,
		Title:         p.Title,
		Email:         p.Email,
		EmailVerified: p.EmailVerified,
		LinkedinURL:   p.LinkedinURL,
	}
}

// A provider can look a person up by LinkedIn URL, or by its own ID for people it found itself.
func (p Person) LookupIdentifier(providerCode string) string {
	if p.HasLinkedin() {
		return p.LinkedinURL
	}

	if p.Source == providerCode {
		return p.ProviderPersonID
	}

	return ""
}

func (p Person) Merge(details *ContactDetails) Person {
	if details == nil {
		return p
	}

	if details.FullName != "" {
		p.FullName = details.FullName
	}

	if details.Title != "" {
		p.Title = details.Title
	}

	if details.LinkedinURL != "" {
		p.LinkedinURL = details.LinkedinURL
	}

	return p
}

func (p Person) ToCompanyContactModel(
	job *jobModel.JobPosting,
	role, providerCode string,
	details *ContactDetails,
	emailStatus modelHelpers.EmailVerificationStatus,
) *enrichmentModel.CompanyContact {
	merged := p.Merge(details)

	contact := enrichmentModel.NewCompanyContact()
	contact.CompanyDomain = job.CompanyDomain
	contact.CompanyName = job.CompanyName
	contact.FullName = strings.TrimSpace(merged.FullName)
	contact.RoleType = modelHelpers.ContactRole(role)
	contact.Title = strings.TrimSpace(merged.Title)
	contact.EnrichmentProvider = providerCode

	if merged.ProviderPersonID != "" {
		contact.ProviderPersonID = &merged.ProviderPersonID
	}

	if linkedinURL := NormalizeLinkedinURL(merged.LinkedinURL); linkedinURL != "" {
		contact.LinkedinURL = &linkedinURL
	}

	if details == nil {
		return contact
	}

	contact.SetEmail(modelHelpers.NormalizeEmail(details.Email), emailStatus)

	if details.Phone != "" {
		contact.PhoneNumber = &details.Phone
	}

	if len(details.Raw) > 0 && json.Valid(details.Raw) && details.Raw[0] == '{' {
		contact.EnrichmentMetadata = details.Raw
	}

	return contact
}

func NormalizeLinkedinURL(linkedinURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(linkedinURL), "/")
	if trimmed == "" {
		return ""
	}

	trimmed = strings.TrimPrefix(trimmed, "http://")
	trimmed = strings.TrimPrefix(trimmed, "https://")

	return "https://" + trimmed
}

func NewRequestEnrichmentResponse(jobID, taskName, status string) *RequestEnrichmentResponse {
	return &RequestEnrichmentResponse{JobID: jobID, TaskName: taskName, Status: status}
}

func ResolveTargetRoles(requested []string, job *jobModel.JobPosting) []string {
	if len(requested) > 0 {
		return requested
	}

	department := ""
	if job.Department != nil {
		department = *job.Department
	}

	return config.GetDefaultTargetRoles(job.JobTitle, department)
}

// PersonFromContact describes a stored contact the way providers are asked about people.
func PersonFromContact(contact *enrichmentModel.CompanyContact) Person {
	person := Person{
		Source:      contact.EnrichmentProvider,
		FullName:    contact.FullName,
		Title:       contact.Title,
		CompanyName: contact.CompanyName,
	}

	if contact.ProviderPersonID != nil {
		person.ProviderPersonID = *contact.ProviderPersonID
	}

	if contact.LinkedinURL != nil {
		person.LinkedinURL = *contact.LinkedinURL
	}

	return person
}
