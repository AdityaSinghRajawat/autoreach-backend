package applicationManagement

import "github.com/AdityaSinghRajawat/autoreach-backend/config"

type ApplicationProfile struct {
	Phone               string            `json:"phone"               validate:"omitempty,max=50"`
	Location            string            `json:"location"            validate:"omitempty,max=255"`
	LinkedinURL         string            `json:"linkedinUrl"         validate:"omitempty,url,max=500"`
	GithubURL           string            `json:"githubUrl"           validate:"omitempty,url,max=500"`
	WebsiteURL          string            `json:"websiteUrl"          validate:"omitempty,url,max=500"`
	CurrentCompany      string            `json:"currentCompany"      validate:"omitempty,max=255"`
	WorkAuthorized      *bool             `json:"workAuthorized"`
	RequiresSponsorship *bool             `json:"requiresSponsorship"`
	SalaryExpectation   string            `json:"salaryExpectation"   validate:"omitempty,max=255"`
	NoticePeriod        string            `json:"noticePeriod"        validate:"omitempty,max=255"`
	HowDidYouHear       string            `json:"howDidYouHear"       validate:"omitempty,max=255"`
	CoverLetter         string            `json:"coverLetter"         validate:"omitempty,max=10000"`
	Answers             map[string]string `json:"answers"             validate:"omitempty,max=100,dive,keys,max=500,endkeys,max=5000"`
}

func (p ApplicationProfile) FieldValue(field string) string {
	switch field {
	case config.ProfileFieldLinkedin:
		return p.LinkedinURL
	case config.ProfileFieldGithub:
		return p.GithubURL
	case config.ProfileFieldWebsite:
		return p.WebsiteURL
	case config.ProfileFieldWorkAuth:
		return yesNo(p.WorkAuthorized)
	case config.ProfileFieldSponsorship:
		return yesNo(p.RequiresSponsorship)
	case config.ProfileFieldSalary:
		return p.SalaryExpectation
	case config.ProfileFieldNoticePeriod:
		return p.NoticePeriod
	case config.ProfileFieldHeardFrom:
		return p.HowDidYouHear
	case config.ProfileFieldCurrentEmployer:
		return p.CurrentCompany
	case config.ProfileFieldLocation:
		return p.Location
	default:
		return ""
	}
}

func yesNo(value *bool) string {
	switch {
	case value == nil:
		return ""
	case *value:
		return config.GetAnswerYes()
	default:
		return config.GetAnswerNo()
	}
}
