CREATE TABLE IF NOT EXISTS company_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_domain VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role_type contact_role NOT NULL,
    title VARCHAR(255) NOT NULL,
    work_email VARCHAR(255),
    phone_number VARCHAR(50),
    linkedin_url VARCHAR(500),
    enrichment_provider VARCHAR(50) NOT NULL,
    provider_person_id VARCHAR(255),
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verification_status VARCHAR(20) NOT NULL DEFAULT 'UNVERIFIED',
    enrichment_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_company_contacts_company_role ON company_contacts (company_domain, role_type);
CREATE UNIQUE INDEX idx_company_contacts_company_email ON company_contacts (company_domain, LOWER(work_email))
    WHERE work_email IS NOT NULL;
CREATE UNIQUE INDEX idx_company_contacts_linkedin ON company_contacts (LOWER(linkedin_url))
    WHERE linkedin_url IS NOT NULL;
CREATE UNIQUE INDEX idx_company_contacts_provider_person ON company_contacts (enrichment_provider, provider_person_id)
    WHERE provider_person_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS job_contacts (
    job_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES company_contacts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (job_id, contact_id)
);

CREATE INDEX idx_job_contacts_contact ON job_contacts (contact_id);

CREATE TABLE IF NOT EXISTS enrichment_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES company_contacts(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    external_request_id VARCHAR(255) NOT NULL,
    credits_used INTEGER NOT NULL DEFAULT 0 CHECK (credits_used >= 0),
    failure_reason TEXT,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_enrichment_request_external UNIQUE (provider, external_request_id)
);

CREATE INDEX idx_enrichment_requests_pending ON enrichment_requests (created_at) WHERE status = 'PENDING';
