package utils

import (
	"regexp"
	"strings"
)

type robotsRule struct {
	pattern *regexp.Regexp
	length  int
	allow   bool
}

// RobotsRules are the robots.txt rules that apply to one user agent (RFC 9309).
type RobotsRules struct {
	rules []robotsRule
}

type robotsGroup struct {
	agents []string
	rules  []robotsRule
}

func ParseRobots(body []byte, userAgent string) RobotsRules {
	groups := []*robotsGroup{}

	var current *robotsGroup

	lastWasAgent := false

	for _, line := range strings.Split(string(body), "\n") {
		line, _, _ = strings.Cut(line, "#")

		key, value, found := strings.Cut(strings.TrimSpace(line), ":")
		if !found {
			continue
		}

		key, value = strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)

		switch key {
		case "user-agent":
			if current == nil || !lastWasAgent {
				current = &robotsGroup{}
				groups = append(groups, current)
			}

			current.agents = append(current.agents, strings.ToLower(value))
			lastWasAgent = true
		case "allow", "disallow":
			lastWasAgent = false

			if current == nil || value == "" {
				continue
			}

			current.rules = append(current.rules, robotsRule{
				pattern: robotsPattern(value),
				length:  len(value),
				allow:   key == "allow",
			})
		default:
			lastWasAgent = false
		}
	}

	return RobotsRules{rules: rulesFor(groups, robotsProductToken(userAgent))}
}

func (r RobotsRules) Allows(path string) bool {
	if path == "" {
		path = "/"
	}

	matched, allowed := -1, true

	for _, rule := range r.rules {
		if !rule.pattern.MatchString(path) {
			continue
		}

		if rule.length > matched || (rule.length == matched && rule.allow) {
			matched, allowed = rule.length, rule.allow
		}
	}

	return allowed
}

func DisallowAllRobots() RobotsRules {
	return RobotsRules{rules: []robotsRule{{pattern: regexp.MustCompile(`^/`), length: 1}}}
}

func rulesFor(groups []*robotsGroup, productToken string) []robotsRule {
	specific, wildcard := []robotsRule{}, []robotsRule{}
	foundSpecific := false

	for _, group := range groups {
		for _, agent := range group.agents {
			switch {
			case agent == "*":
				wildcard = append(wildcard, group.rules...)
			case productToken != "" && strings.Contains(productToken, agent):
				specific = append(specific, group.rules...)
				foundSpecific = true
			}
		}
	}

	if foundSpecific {
		return specific
	}

	return wildcard
}

// "AutoReachBot/1.0" is matched against robots.txt groups as "autoreachbot".
func robotsProductToken(userAgent string) string {
	token, _, _ := strings.Cut(strings.ToLower(userAgent), "/")

	return strings.TrimSpace(token)
}

// Robots patterns are path prefixes where * matches anything and a trailing $ anchors the end.
func robotsPattern(value string) *regexp.Regexp {
	anchored := strings.HasSuffix(value, "$")
	value = strings.TrimSuffix(value, "$")

	expression := "^" + strings.ReplaceAll(regexp.QuoteMeta(value), `\*`, ".*")
	if anchored {
		expression += "$"
	}

	return regexp.MustCompile(expression)
}
