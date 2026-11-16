package config

import (
	"log"
	"regexp"
	"strings"
)

type Table string

const (
	TableUsers                 Table = "users"
	TableRefreshTokens         Table = "refresh_tokens"
	TableCandidates            Table = "candidates"
	TableAPIClients            Table = "api_clients"
	TableJobPostings           Table = "job_postings"
	TableCandidateResumes      Table = "candidate_resumes"
	TableJobMatches            Table = "job_matches"
	TableCompanyContacts       Table = "company_contacts"
	TableJobContacts           Table = "job_contacts"
	TableEnrichmentRequests    Table = "enrichment_requests"
	TableChannelAccounts       Table = "candidate_channel_accounts"
	TableOutreachTemplates     Table = "outreach_templates"
	TableOutreachCampaigns     Table = "outreach_campaigns"
	TableOutreachCampaignSteps Table = "outreach_campaign_steps"
	TableOutreachMessages      Table = "outreach_messages"
	TableOutreachEvents        Table = "outreach_events"
	TableSuppressionEntries    Table = "suppression_entries"
	TableApplications          Table = "applications"
	TableJobSources            Table = "job_sources"
	TableCompanyDomains        Table = "company_domains"
	TableContactSignals        Table = "contact_signals"
	TableActivityEvents        Table = "activity_events"
)

var declaredTables = []Table{
	TableUsers,
	TableRefreshTokens,
	TableCandidates,
	TableAPIClients,
	TableJobPostings,
	TableCandidateResumes,
	TableJobMatches,
	TableCompanyContacts,
	TableJobContacts,
	TableEnrichmentRequests,
	TableChannelAccounts,
	TableOutreachTemplates,
	TableOutreachCampaigns,
	TableOutreachCampaignSteps,
	TableOutreachMessages,
	TableOutreachEvents,
	TableSuppressionEntries,
	TableApplications,
	TableJobSources,
	TableCompanyDomains,
	TableContactSignals,
}

// declaredViews are read-only views; their names can be overridden like tables, but they hold no rows.
var declaredViews = []Table{
	TableActivityEvents,
}

var (
	tableNamesI            = map[Table]string{}
	tableIdentifierPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
)

func loadTableNames() {
	for _, table := range append(GetDeclaredTables(), declaredViews...) {
		name := getEnvWithDefault("DB_TABLE_"+strings.ToUpper(string(table)), string(table))
		if !tableIdentifierPattern.MatchString(name) {
			log.Fatalf("table name override for %s is not a valid identifier", table)
		}

		tableNamesI[table] = name
	}
}

func GetDeclaredTables() []Table {
	return append([]Table(nil), declaredTables...)
}

func GetTableName(table Table) string {
	name, found := tableNamesI[table]
	if !found {
		return string(table)
	}

	return name
}
