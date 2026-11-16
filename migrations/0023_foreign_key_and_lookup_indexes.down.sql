CREATE INDEX IF NOT EXISTS idx_outreach_messages_contact ON outreach_messages (candidate_id, contact_id);

DROP INDEX IF EXISTS idx_company_contacts_email;
DROP INDEX IF EXISTS idx_job_postings_job_source_any;
DROP INDEX IF EXISTS idx_enrichment_requests_job;
DROP INDEX IF EXISTS idx_enrichment_requests_contact;
DROP INDEX IF EXISTS idx_job_matches_resume;
DROP INDEX IF EXISTS idx_applications_job;
DROP INDEX IF EXISTS idx_outreach_messages_email_sent;
DROP INDEX IF EXISTS idx_outreach_messages_job;
DROP INDEX IF EXISTS idx_outreach_messages_campaign;
DROP INDEX IF EXISTS idx_outreach_messages_contact_created;
