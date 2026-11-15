CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    resume_id UUID REFERENCES candidate_resumes(id) ON DELETE SET NULL,
    match_score NUMERIC(5,4) NOT NULL CHECK (match_score >= 0 AND match_score <= 1),
    status application_status NOT NULL DEFAULT 'QUEUED',
    ats_type VARCHAR(20) NOT NULL,
    apply_url VARCHAR(1000) NOT NULL,
    prepared_answers JSONB NOT NULL DEFAULT '[]'::jsonb,
    missing_answers JSONB NOT NULL DEFAULT '[]'::jsonb,
    execution_logs JSONB NOT NULL DEFAULT '[]'::jsonb,
    external_confirmation VARCHAR(255),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    failure_reason TEXT,
    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_application_candidate_job UNIQUE (candidate_id, job_id)
);

CREATE INDEX idx_applications_candidate ON applications (candidate_id, created_at DESC);
CREATE INDEX idx_applications_pending ON applications (updated_at) WHERE status IN ('QUEUED', 'IN_PROGRESS');
