package config

import "strings"

const (
	ContactRoleEngineeringManager = "ENGINEERING_MANAGER"
	ContactRoleHiringManager      = "HIRING_MANAGER"
	ContactRoleRecruiter          = "RECRUITER"
)

var roleClassificationOrderI = []string{
	ContactRoleRecruiter,
	ContactRoleEngineeringManager,
	ContactRoleHiringManager,
}

var roleTitleKeywordsI = map[string][]string{
	ContactRoleEngineeringManager: {
		"engineering manager", "head of engineering", "director of engineering",
		"vp of engineering", "vp engineering", "vice president of engineering",
		"tech lead", "technical lead", "engineering lead", "chief technology officer", "cto",
	},
	ContactRoleRecruiter: {
		"recruiter", "recruiting", "talent acquisition", "talent partner", "sourcer",
		"people partner", "head of talent", "head of people",
	},
	ContactRoleHiringManager: {
		"hiring manager", "head of", "director", "vice president", "vp", "department head",
		"general manager", "manager",
	},
}

var engineeringKeywordsI = []string{
	"engineer", "developer", "software", "devops", "sre", "backend", "frontend", "full stack",
	"fullstack", "platform", "infrastructure", "machine learning", "data scientist", "architect",
	"security", "qa", "sdet", "programmer",
}

var roleLocalPartsI = map[string]struct{}{
	"info": {}, "hr": {}, "jobs": {}, "careers": {}, "recruiting": {}, "talent": {},
	"support": {}, "admin": {}, "contact": {}, "hello": {}, "sales": {}, "team": {},
	"noreply": {}, "no-reply": {}, "office": {}, "help": {},
}

var disposableDomainsI = map[string]struct{}{
	"mailinator.com": {}, "guerrillamail.com": {}, "10minutemail.com": {}, "tempmail.com": {},
	"yopmail.com": {}, "trashmail.com": {}, "sharklasers.com": {}, "getnada.com": {},
}

func GetSearchTitles(roles []string) []string {
	titles := make([]string, 0, len(roles)*8)
	for _, role := range roles {
		titles = append(titles, roleTitleKeywordsI[role]...)
	}

	return titles
}

func ClassifyContactRole(title string) (string, bool) {
	words := padWords(title)

	for _, role := range roleClassificationOrderI {
		for _, keyword := range roleTitleKeywordsI[role] {
			if strings.Contains(words, padWords(keyword)) {
				return role, true
			}
		}
	}

	return "", false
}

func GetDefaultTargetRoles(jobTitle, department string) []string {
	words := padWords(jobTitle + " " + department)

	for _, keyword := range engineeringKeywordsI {
		if strings.Contains(words, " "+strings.TrimSpace(keyword)) {
			return []string{ContactRoleEngineeringManager, ContactRoleRecruiter}
		}
	}

	return []string{ContactRoleHiringManager, ContactRoleRecruiter}
}

func padWords(text string) string {
	normalised := strings.Map(func(character rune) rune {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			return character
		case character >= 'A' && character <= 'Z':
			return character + ('a' - 'A')
		default:
			return ' '
		}
	}, text)

	return " " + strings.Join(strings.Fields(normalised), " ") + " "
}

func IsRoleLocalPart(localPart string) bool {
	_, found := roleLocalPartsI[strings.ToLower(localPart)]

	return found
}

func IsDisposableDomain(domain string) bool {
	_, found := disposableDomainsI[strings.ToLower(domain)]

	return found
}
