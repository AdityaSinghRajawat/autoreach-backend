-- The activity feed is a read-side view over the tables that already record each step, so every
-- state change appears without the services writing a second copy.
CREATE OR REPLACE VIEW activity_events AS
SELECT 'match:' || m.id::text AS event_id,
       m.candidate_id,
       'JOB_MATCHED'::text AS event_type,
       NULL::text AS channel,
       m.created_at AS occurred_at,
       m.job_id,
       NULL::uuid AS contact_id,
       NULL::uuid AS application_id,
       NULL::uuid AS message_id
FROM job_matches m
JOIN candidates c ON c.id = m.candidate_id
WHERE m.match_score >= c.min_match_score
UNION ALL
SELECT 'application:' || a.id::text || ':SUBMITTED', a.candidate_id, 'APPLICATION_SUBMITTED', NULL,
       a.submitted_at, a.job_id, NULL, a.id, NULL
FROM applications a
WHERE a.submitted_at IS NOT NULL
UNION ALL
SELECT 'application:' || a.id::text || ':' || a.status::text, a.candidate_id,
       'APPLICATION_' || a.status::text, NULL, a.updated_at, a.job_id, NULL, a.id, NULL
FROM applications a
WHERE a.status IN ('ACTION_REQUIRED', 'INTERVIEWING', 'REJECTED')
UNION ALL
SELECT 'outreach:' || e.id::text, m.candidate_id, 'OUTREACH_' || e.event_type, m.channel::text,
       e.occurred_at, m.job_id, m.contact_id, NULL, m.id
FROM outreach_events e
JOIN outreach_messages m ON m.id = e.message_id;
