CREATE TABLE IF NOT EXISTS job_postings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_hash VARCHAR(64) NOT NULL UNIQUE,
    source job_source NOT NULL DEFAULT 'DIRECT',
    job_source_id UUID REFERENCES job_sources(id) ON DELETE SET NULL,
    company_name VARCHAR(255) NOT NULL,
    company_domain VARCHAR(255) NOT NULL,
    job_title VARCHAR(255) NOT NULL,
    normalized_title VARCHAR(255) NOT NULL,
    job_description TEXT NOT NULL,
    description_hash VARCHAR(64) NOT NULL,
    location VARCHAR(255),
    department VARCHAR(255),
    salary_min NUMERIC(12,2),
    salary_max NUMERIC(12,2),
    salary_currency VARCHAR(3),
    ats_type VARCHAR(50) NOT NULL,
    external_url VARCHAR(1000) NOT NULL,
    external_job_id VARCHAR(255),
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    embedding vector(1536),
    embedding_model VARCHAR(100),
    embedded_at TIMESTAMPTZ,
    status job_status NOT NULL DEFAULT 'ACTIVE',
    posted_at TIMESTAMPTZ,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_job_postings_salary_range CHECK (salary_min IS NULL OR salary_max IS NULL OR salary_min <= salary_max)
);

CREATE INDEX idx_job_postings_status_created ON job_postings (status, created_at DESC);
CREATE INDEX idx_job_postings_company_domain ON job_postings (company_domain);
CREATE INDEX idx_job_postings_job_source ON job_postings (job_source_id, last_seen_at) WHERE status = 'ACTIVE';
CREATE INDEX idx_job_postings_job_source_any ON job_postings (job_source_id) WHERE job_source_id IS NOT NULL;
CREATE INDEX idx_job_postings_pending_embedding ON job_postings (updated_at) WHERE embedding IS NULL AND status = 'ACTIVE';

CREATE INDEX idx_job_postings_embedding ON job_postings
USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)
WHERE status = 'ACTIVE';
