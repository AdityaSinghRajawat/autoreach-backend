CREATE TABLE IF NOT EXISTS outreach_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES outreach_messages(id) ON DELETE CASCADE,
    event_type VARCHAR(20) NOT NULL
        CHECK (event_type IN ('SENT', 'OPENED', 'CLICKED', 'BOUNCED', 'REPLIED', 'AUTO_REPLIED')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outreach_events_message ON outreach_events (message_id, occurred_at);
