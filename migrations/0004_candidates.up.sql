CREATE TABLE IF NOT EXISTS candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    linkedin_profile_url VARCHAR(500),
    min_match_score NUMERIC(3,2) NOT NULL DEFAULT 0.80 CHECK (min_match_score >= 0.00 AND min_match_score <= 1.00),
    auto_apply_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    auto_outreach_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_candidates_email_lower ON candidates (LOWER(email));
