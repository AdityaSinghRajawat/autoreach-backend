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
