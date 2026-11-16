package config

// GetUUIDPathParams lists the path parameters that hold UUIDs; a malformed value answers 404.
// Queue task IDs are not listed because the queue chooses their format.
func GetUUIDPathParams() []string {
	return []string{
		GetCandidateIDParam(),
		GetJobIDParam(),
		GetClientIDParam(),
		GetResumeIDParam(),
		GetAccountIDParam(),
		GetTemplateIDParam(),
		GetCampaignIDParam(),
		GetMessageIDParam(),
		GetApplicationIDParam(),
		GetJobSourceIDParam(),
		GetSuppressionIDParam(),
		GetContactIDParam(),
		GetContactSignalIDParam(),
	}
}
