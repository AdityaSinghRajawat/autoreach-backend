CREATE TABLE IF NOT EXISTS candidate_resumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    storage_key VARCHAR(500) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    content_sha256 VARCHAR(64) NOT NULL,
    raw_text TEXT NOT NULL DEFAULT '',
    parsed_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    parse_status resume_status NOT NULL DEFAULT 'UPLOADED',
    failure_reason TEXT,
    embedding vector(1536),
    embedding_model VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidate_resumes_candidate ON candidate_resumes (candidate_id, created_at DESC);
CREATE UNIQUE INDEX idx_candidate_resumes_one_active ON candidate_resumes (candidate_id) WHERE is_active;
CREATE INDEX idx_candidate_resumes_pending ON candidate_resumes (updated_at)
    WHERE parse_status IN ('UPLOADED', 'PARSED');

CREATE INDEX idx_candidate_resumes_embedding ON candidate_resumes
USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
