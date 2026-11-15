CREATE TABLE IF NOT EXISTS job_contacts (
    job_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES company_contacts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (job_id, contact_id)
);

CREATE INDEX idx_job_contacts_contact ON job_contacts (contact_id);
