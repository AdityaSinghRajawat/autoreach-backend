package outreachManagement

import (
	"regexp"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type MessageContext struct {
	Candidate *candidateModel.Candidate
	Job       *jobModel.JobPosting
	Contact   *enrichmentModel.CompanyContact
	Skills    []string
}

func (c MessageContext) PlaceholderValues() map[string]string {
	values := map[string]string{
		config.PlaceholderCandidateName:      c.Candidate.FullName,
		config.PlaceholderCandidateFirstName: utils.GetFirstName(c.Candidate.FullName),
		config.PlaceholderContactName:        c.Contact.FullName,
		config.PlaceholderContactFirstName:   utils.GetFirstName(c.Contact.FullName),
		config.PlaceholderContactTitle:       c.Contact.Title,
		config.PlaceholderCompanyName:        c.Job.CompanyName,
		config.PlaceholderJobTitle:           c.Job.JobTitle,
		config.PlaceholderMatchingSkills:     utils.JoinWithAnd(c.matchingSkills()),
	}

	if c.Candidate.LinkedinProfileURL != nil {
		values[config.PlaceholderCandidateLinkedin] = *c.Candidate.LinkedinProfileURL
	}

	return values
}

func (c MessageContext) matchingSkills() []string {
	description := " " + strings.ToLower(c.Job.JobDescription) + " "
	matching := []string{}

	for _, skill := range c.Skills {
		pattern := `[^a-z0-9+#.]` + regexp.QuoteMeta(skill) + `[^a-z0-9+#]`
		if matched, _ := regexp.MatchString(pattern, description); matched {
			matching = append(matching, skill)
		}

		if len(matching) == config.GetMaxMatchingSkillsInMessage() {
			break
		}
	}

	return matching
}
