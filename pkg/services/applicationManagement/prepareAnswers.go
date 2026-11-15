package applicationManagement

import (
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func prepareAnswers(
	form applicationTypes.ApplicationForm,
	profile applicationTypes.ApplicationProfile,
	candidate *candidateModel.Candidate,
	hasResume bool,
) ([]applicationTypes.PreparedAnswer, []applicationTypes.MissingAnswer) {
	answers := []applicationTypes.PreparedAnswer{}
	missing := []applicationTypes.MissingAnswer{}

	for _, question := range form.Questions {
		answered := false

		for _, field := range question.Fields {
			answer, found := answerField(question.Label, field, profile, candidate, hasResume)
			if found {
				answers = append(answers, answer)
				answered = true
			}
		}

		if question.Required && !answered && len(question.Fields) > 0 {
			missing = append(missing, applicationTypes.MissingAnswer{
				Label:     question.Label,
				FieldName: question.Fields[0].Name,
			})
		}
	}

	return answers, missing
}

func answerField(
	label string,
	field applicationTypes.FormField,
	profile applicationTypes.ApplicationProfile,
	candidate *candidateModel.Candidate,
	hasResume bool,
) (applicationTypes.PreparedAnswer, bool) {
	text := knownFieldValue(field.Name, profile, candidate, hasResume)
	if text == "" && field.Type != "input_file" && field.Type != "input_hidden" {
		text = labelValue(label, profile, candidate)
	}

	if text == "" {
		return applicationTypes.PreparedAnswer{}, false
	}

	answer := applicationTypes.PreparedAnswer{Label: label, FieldName: field.Name, Value: text}
	if len(field.Options) == 0 {
		return answer, true
	}

	for _, option := range field.Options {
		if strings.EqualFold(strings.TrimSpace(option.Label), text) {
			answer.Value = option.Value
			if field.Type == "multi_value_multi_select" {
				answer.Values = []string{option.Value}
			}

			return answer, true
		}
	}

	return applicationTypes.PreparedAnswer{}, false
}

func knownFieldValue(
	name string,
	profile applicationTypes.ApplicationProfile,
	candidate *candidateModel.Candidate,
	hasResume bool,
) string {
	firstName, lastName := splitName(candidate.FullName)

	switch name {
	case "first_name":
		return firstName
	case "last_name":
		return lastName
	case "name":
		return candidate.FullName
	case "email":
		return candidate.Email
	case "phone":
		return profile.Phone
	case "org":
		return profile.CurrentCompany
	case "location":
		return profile.Location
	case "cover_letter_text", "comments":
		return profile.CoverLetter
	case config.GetLeverLinkedinURLField():
		return linkedinURL(profile, candidate)
	case config.GetLeverGithubURLField():
		return profile.GithubURL
	case config.GetLeverPortfolioURLField():
		return profile.WebsiteURL
	case "resume":
		if hasResume {
			return config.GetAnswerYes()
		}
	}

	return ""
}

func labelValue(
	label string,
	profile applicationTypes.ApplicationProfile,
	candidate *candidateModel.Candidate,
) string {
	for question, answer := range profile.Answers {
		if strings.EqualFold(strings.TrimSpace(question), strings.TrimSpace(label)) {
			return answer
		}
	}

	field := config.ProfileFieldForLabel(label)
	if field == config.ProfileFieldLinkedin {
		return linkedinURL(profile, candidate)
	}

	return profile.FieldValue(field)
}

func linkedinURL(
	profile applicationTypes.ApplicationProfile,
	candidate *candidateModel.Candidate,
) string {
	if profile.LinkedinURL != "" || candidate.LinkedinProfileURL == nil {
		return profile.LinkedinURL
	}

	return *candidate.LinkedinProfileURL
}

func splitName(fullName string) (string, string) {
	fields := strings.Fields(fullName)
	if len(fields) == 0 {
		return "", ""
	}

	return fields[0], strings.Join(fields[1:], " ")
}
