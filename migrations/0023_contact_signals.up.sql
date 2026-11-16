CREATE TABLE IF NOT EXISTS contact_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL REFERENCES company_contacts(id) ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    signal_type VARCHAR(30) NOT NULL CHECK (signal_type IN ('BOUNCED', 'REPLIED', 'LINKEDIN_RESPONDED')),
    address VARCHAR(255) NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    applied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_contact_signal_reporter UNIQUE (contact_id, signal_type, address, candidate_id)
);

CREATE INDEX idx_contact_signals_candidate ON contact_signals (candidate_id);
