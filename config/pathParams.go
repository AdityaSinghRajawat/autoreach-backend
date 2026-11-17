package config

// Queue task IDs are absent: the queue chooses their format.
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
