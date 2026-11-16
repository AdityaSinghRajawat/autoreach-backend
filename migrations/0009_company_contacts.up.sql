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
    last_interaction_at TIMESTAMPTZ,
    email_confirmed_at TIMESTAMPTZ,
    linkedin_confirmed_at TIMESTAMPTZ,
    anonymized_at TIMESTAMPTZ,
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
