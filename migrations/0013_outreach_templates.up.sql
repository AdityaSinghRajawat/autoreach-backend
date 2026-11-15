CREATE TABLE IF NOT EXISTS outreach_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    channel outreach_channel NOT NULL,
    subject VARCHAR(500),
    body TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'ARCHIVED')),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outreach_templates_candidate ON outreach_templates (candidate_id, created_at DESC);
