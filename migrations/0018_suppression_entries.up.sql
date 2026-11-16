CREATE TABLE IF NOT EXISTS suppression_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID REFERENCES candidates(id) ON DELETE CASCADE,
    entry_type VARCHAR(10) NOT NULL CHECK (entry_type IN ('EMAIL', 'DOMAIN')),
    value VARCHAR(255) NOT NULL,
    reason VARCHAR(20) NOT NULL CHECK (reason IN ('BOUNCED', 'UNSUBSCRIBED', 'COMPLAINT', 'MANUAL', 'ERASED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_suppression_global ON suppression_entries (entry_type, value) WHERE candidate_id IS NULL;
CREATE UNIQUE INDEX idx_suppression_candidate ON suppression_entries (candidate_id, entry_type, value)
    WHERE candidate_id IS NOT NULL;
