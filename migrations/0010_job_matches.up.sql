CREATE TABLE IF NOT EXISTS job_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    resume_id UUID REFERENCES candidate_resumes(id) ON DELETE SET NULL,
    raw_similarity NUMERIC(6,5) NOT NULL CHECK (raw_similarity >= -1 AND raw_similarity <= 1),
    match_score NUMERIC(5,4) NOT NULL CHECK (match_score >= 0.0000 AND match_score <= 1.0000),
    fanned_out_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_job_match_candidate_job UNIQUE (candidate_id, job_id)
);

CREATE INDEX idx_job_matches_candidate_score ON job_matches (candidate_id, match_score DESC, created_at DESC);
CREATE INDEX idx_job_matches_job ON job_matches (job_id);
CREATE INDEX idx_job_matches_resume ON job_matches (resume_id);
