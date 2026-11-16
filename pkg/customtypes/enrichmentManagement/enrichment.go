package enrichmentManagement

import (
	"encoding/json"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type DiscoveryQuery struct {
	CompanyDomain string
	Titles        []string
	Limit         int
}

type Person struct {
	ProviderPersonID string
	FullName         string
	Title            string
	LinkedinURL      string
	CompanyName      string
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

func (p Person) HasLinkedin() bool {
	return p.LinkedinURL != ""
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
