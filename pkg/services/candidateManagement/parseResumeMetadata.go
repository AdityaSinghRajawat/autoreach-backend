package candidateManagement

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *CandidateService) parseResumeMetadata(
	ctx context.Context,
	text string,
) candidateTypes.ResumeMetadata {
	metadata := candidateTypes.ResumeMetadata{
		WordCount: len(strings.Fields(text)),
		ParsedBy:  s.chatCompleter.ChatModel(),
	}

	profile, err := s.extractResumeProfile(ctx, text)
	if err != nil {
		utils.LogWarn(ctx, "resume parsing model unavailable, using the skill dictionary",
			"error", err.Error())

		profile = candidateTypes.ResumeProfile{Skills: matchDictionarySkills(text)}
		metadata.ParsedBy = config.GetResumeParsedByDictionary()
	}

	metadata.ResumeProfile = normalizeResumeProfile(profile)

	return metadata
}

func (s *CandidateService) extractResumeProfile(
	ctx context.Context,
	text string,
) (candidateTypes.ResumeProfile, error) {
	reply, err := s.chatCompleter.CreateChatCompletion(ctx, sharedTypes.ChatCompletionRequest{
		Messages: []sharedTypes.ChatMessage{
			{Role: config.GetChatRoleSystem(), Content: config.GetResumeParsingInstructions()},
			{
				Role:    config.GetChatRoleUser(),
				Content: truncateRunes(text, config.GetResumeParsingMaxInputChars()),
			},
		},
		ResponseSchema: &sharedTypes.ChatResponseSchema{
			Name:   config.GetResumeProfileSchemaName(),
			Schema: config.GetResumeProfileSchema(),
		},
	})
	if err != nil {
		return candidateTypes.ResumeProfile{}, err
	}

	profile := candidateTypes.ResumeProfile{}
	if err := json.Unmarshal([]byte(reply), &profile); err != nil {
		return candidateTypes.ResumeProfile{}, fmt.Errorf("decode resume profile: %w", err)
	}

	return profile, nil
}

func normalizeResumeProfile(profile candidateTypes.ResumeProfile) candidateTypes.ResumeProfile {
	normalized := candidateTypes.ResumeProfile{
		Skills: uniqueValues(
			profile.Skills,
			strings.ToLower,
			config.GetResumeParsingMaxSkills(),
		),
		JobTitles: uniqueValues(profile.JobTitles, nil, config.GetResumeParsingMaxJobTitles()),
		Education: make([]candidateTypes.ResumeEducation, 0, len(profile.Education)),
	}

	slices.Sort(normalized.Skills)

	if years := profile.YearsOfExperience; years != nil &&
		*years >= 0 && *years <= config.GetResumeParsingMaxYears() {
		normalized.YearsOfExperience = years
	}

	for _, education := range profile.Education {
		if len(normalized.Education) == config.GetResumeParsingMaxEducation() {
			break
		}

		education.Degree = strings.TrimSpace(education.Degree)
		education.FieldOfStudy = strings.TrimSpace(education.FieldOfStudy)
		education.Institution = strings.TrimSpace(education.Institution)

		if education != (candidateTypes.ResumeEducation{}) {
			normalized.Education = append(normalized.Education, education)
		}
	}

	return normalized
}

func uniqueValues(values []string, transform func(string) string, limit int) []string {
	unique := make([]string, 0, min(len(values), limit))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if transform != nil {
			value = transform(value)
		}

		if value != "" && !slices.Contains(unique, value) && len(unique) < limit {
			unique = append(unique, value)
		}
	}

	return unique
}

func matchDictionarySkills(text string) []string {
	lowered := " " + strings.ToLower(text) + " "
	skills := make([]string, 0, 16)

	for _, skill := range config.GetSkillDictionary() {
		pattern := `(^|[^a-z0-9+#.])` + regexp.QuoteMeta(skill) + `($|[^a-z0-9+#])`
		if matched, err := regexp.MatchString(pattern, lowered); err == nil && matched {
			skills = append(skills, skill)
		}
	}

	return skills
}
