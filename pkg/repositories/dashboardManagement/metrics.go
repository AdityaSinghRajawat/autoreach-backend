package dashboardManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
)

type MetricsRepo struct {
	postgres              *db.PostgresConn
	candidatesTable       string
	jobPostingsTable      string
	jobMatchesTable       string
	jobContactsTable      string
	applicationsTable     string
	outreachMessagesTable string
}

func NewMetricsRepo(postgres *db.PostgresConn) (*MetricsRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new metrics repo: %w", config.ErrMissingDependency)
	}

	return &MetricsRepo{
		postgres:              postgres,
		candidatesTable:       config.GetTableName(config.TableCandidates),
		jobPostingsTable:      config.GetTableName(config.TableJobPostings),
		jobMatchesTable:       config.GetTableName(config.TableJobMatches),
		jobContactsTable:      config.GetTableName(config.TableJobContacts),
		applicationsTable:     config.GetTableName(config.TableApplications),
		outreachMessagesTable: config.GetTableName(config.TableOutreachMessages),
	}, nil
}

// CountForCandidate returns nil when the candidate does not exist. A job counts as matched when its
// score reaches the candidate's threshold; a response is a contact who replied on any channel.
func (r *MetricsRepo) CountForCandidate(
	ctx context.Context,
	candidateID string,
) (*dashboardTypes.CandidateCounts, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM `+r.jobPostingsTable+` WHERE status = `+sqlJobStatusActive+`),
			(SELECT COUNT(*) FROM `+r.jobMatchesTable+` m
				WHERE m.candidate_id = c.id AND m.match_score >= c.min_match_score),
			(SELECT COUNT(*) FROM `+r.applicationsTable+`
				WHERE candidate_id = c.id AND submitted_at IS NOT NULL),
			(SELECT COUNT(*) FROM `+r.applicationsTable+`
				WHERE candidate_id = c.id AND status = `+sqlApplicationStatusInterviewing+`),
			(SELECT COUNT(DISTINCT jc.contact_id) FROM `+r.jobMatchesTable+` m
				JOIN `+r.jobContactsTable+` jc ON jc.job_id = m.job_id
				WHERE m.candidate_id = c.id AND m.match_score >= c.min_match_score),
			(SELECT COUNT(*) FROM `+r.outreachMessagesTable+`
				WHERE candidate_id = c.id AND channel = `+sqlOutreachChannelEmail+` AND sent_at IS NOT NULL),
			(SELECT COUNT(*) FROM `+r.outreachMessagesTable+`
				WHERE candidate_id = c.id AND channel = `+sqlOutreachChannelLinkedinConnect+` AND sent_at IS NOT NULL),
			(SELECT COUNT(*) FROM `+r.outreachMessagesTable+`
				WHERE candidate_id = c.id AND channel = `+sqlOutreachChannelLinkedinConnect+` AND status IN (`+sqlOutreachStatusAccepted+`, `+sqlOutreachStatusReplied+`)),
			(SELECT COUNT(DISTINCT contact_id) FROM `+r.outreachMessagesTable+`
				WHERE candidate_id = c.id AND replied_at IS NOT NULL)
		FROM `+r.candidatesTable+` c
		WHERE c.id = $1
	`, candidateID)

	return db.FindOne(row, func(row db.RowScanner) (*dashboardTypes.CandidateCounts, error) {
		counts := dashboardTypes.CandidateCounts{}
		err := row.Scan(
			&counts.TotalJobsScraped,
			&counts.JobsMatched,
			&counts.ApplicationsSubmitted,
			&counts.Interviews,
			&counts.ContactsDiscovered,
			&counts.ColdEmailsSent,
			&counts.LinkedinRequestsSent,
			&counts.LinkedinConnectionsAccepted,
			&counts.RecruiterResponses,
		)

		return &counts, err
	}, "count candidate metrics")
}

// CountEmailDeliveries counts cold emails sent since the given time, and how many of them bounced.
func (r *MetricsRepo) CountEmailDeliveries(ctx context.Context, since time.Time) (int, int, error) {
	var sent, bounced int

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = `+sqlOutreachStatusBounced+`)
		FROM `+r.outreachMessagesTable+`
		WHERE channel = `+sqlOutreachChannelEmail+` AND sent_at >= $1
	`, since).Scan(&sent, &bounced)
	if err != nil {
		return 0, 0, fmt.Errorf("count email deliveries: %w", err)
	}

	return sent, bounced, nil
}
