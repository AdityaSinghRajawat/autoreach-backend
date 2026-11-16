-- Indexes for foreign keys (deletes cascade through them) and for lookups that scanned whole tables.
CREATE INDEX IF NOT EXISTS idx_outreach_messages_contact_created ON outreach_messages (contact_id, created_at);
CREATE INDEX IF NOT EXISTS idx_outreach_messages_campaign ON outreach_messages (campaign_id) WHERE campaign_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_outreach_messages_job ON outreach_messages (job_id);
CREATE INDEX IF NOT EXISTS idx_outreach_messages_email_sent ON outreach_messages (sent_at) WHERE channel = 'EMAIL';
CREATE INDEX IF NOT EXISTS idx_applications_job ON applications (job_id);
CREATE INDEX IF NOT EXISTS idx_job_matches_resume ON job_matches (resume_id);
CREATE INDEX IF NOT EXISTS idx_enrichment_requests_contact ON enrichment_requests (contact_id);
CREATE INDEX IF NOT EXISTS idx_enrichment_requests_job ON enrichment_requests (job_id);
CREATE INDEX IF NOT EXISTS idx_job_postings_job_source_any ON job_postings (job_source_id) WHERE job_source_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_company_contacts_email ON company_contacts (LOWER(work_email)) WHERE work_email IS NOT NULL;

-- (candidate_id, contact_id) is already the leading part of the unique key on outreach steps.
DROP INDEX IF EXISTS idx_outreach_messages_contact;
