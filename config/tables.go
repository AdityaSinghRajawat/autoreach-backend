package config

import (
	"log"
	"regexp"
	"strings"
)

type Table string

const (
	TableUsers            Table = "users"
	TableRefreshTokens    Table = "refresh_tokens"
	TableCandidates       Table = "candidates"
	TableAPIClients       Table = "api_clients"
	TableJobPostings      Table = "job_postings"
	TableCandidateResumes Table = "candidate_resumes"
	TableJobMatches       Table = "job_matches"
)

var declaredTables = []Table{
	TableUsers,
	TableRefreshTokens,
	TableCandidates,
	TableAPIClients,
	TableJobPostings,
	TableCandidateResumes,
	TableJobMatches,
}

var (
	tableNamesI            = map[Table]string{}
	tableIdentifierPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
)

func loadTableNames() {
	for _, table := range declaredTables {
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
