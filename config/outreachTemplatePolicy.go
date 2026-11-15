package config

const (
	templateIDParam   = "templateId"
	campaignIDParam   = "campaignId"
	maxMatchingSkills = 5
)

const (
	PlaceholderCandidateName      = "candidateName"
	PlaceholderCandidateFirstName = "candidateFirstName"
	PlaceholderCandidateLinkedin  = "candidateLinkedinUrl"
	PlaceholderContactName        = "contactName"
	PlaceholderContactFirstName   = "contactFirstName"
	PlaceholderContactTitle       = "contactTitle"
	PlaceholderCompanyName        = "companyName"
	PlaceholderJobTitle           = "jobTitle"
	PlaceholderMatchingSkills     = "matchingSkills"
)

var templatePlaceholdersI = []string{
	PlaceholderCandidateName,
	PlaceholderCandidateFirstName,
	PlaceholderCandidateLinkedin,
	PlaceholderContactName,
	PlaceholderContactFirstName,
	PlaceholderContactTitle,
	PlaceholderCompanyName,
	PlaceholderJobTitle,
	PlaceholderMatchingSkills,
}

func GetTemplateIDParam() string         { return templateIDParam }
func GetCampaignIDParam() string         { return campaignIDParam }
func GetMaxMatchingSkillsInMessage() int { return maxMatchingSkills }
func GetTemplatePlaceholders() []string  { return append([]string(nil), templatePlaceholdersI...) }
