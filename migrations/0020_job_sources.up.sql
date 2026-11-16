CREATE TABLE IF NOT EXISTS job_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ats_type VARCHAR(30) NOT NULL CHECK (ats_type IN ('GREENHOUSE', 'LEVER', 'ASHBY', 'SMARTRECRUITERS')),
    board_token VARCHAR(100) NOT NULL,
    region VARCHAR(10) NOT NULL DEFAULT 'GLOBAL' CHECK (region IN ('GLOBAL', 'EU')),
    company_name VARCHAR(255) NOT NULL,
    company_domain VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_polled_at TIMESTAMPTZ,
    last_job_count INTEGER,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_job_source_board UNIQUE (ats_type, board_token)
);

CREATE INDEX idx_job_sources_active ON job_sources (last_polled_at NULLS FIRST) WHERE is_active;

ALTER TABLE job_postings ADD COLUMN job_source_id UUID REFERENCES job_sources(id) ON DELETE SET NULL;
CREATE INDEX idx_job_postings_job_source ON job_postings (job_source_id, last_seen_at) WHERE status = 'ACTIVE';
