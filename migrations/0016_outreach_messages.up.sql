CREATE TABLE IF NOT EXISTS outreach_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES company_contacts(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    campaign_id UUID REFERENCES outreach_campaigns(id) ON DELETE SET NULL,
    channel_account_id UUID REFERENCES candidate_channel_accounts(id) ON DELETE SET NULL,
    template_id UUID REFERENCES outreach_templates(id) ON DELETE SET NULL,
    step_number INTEGER NOT NULL DEFAULT 1 CHECK (step_number > 0),
    channel outreach_channel NOT NULL,
    status outreach_status NOT NULL DEFAULT 'PENDING',
    subject VARCHAR(500),
    body_content TEXT NOT NULL,
    linkedin_note VARCHAR(300),
    scheduled_at TIMESTAMPTZ NOT NULL,
    provider_message_id VARCHAR(255),
    in_reply_to VARCHAR(255),
    tracking_token VARCHAR(64),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    sent_at TIMESTAMPTZ,
    replied_at TIMESTAMPTZ,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_outreach_message_step UNIQUE (candidate_id, contact_id, job_id, channel, step_number)
);

CREATE INDEX idx_outreach_messages_due ON outreach_messages (scheduled_at) WHERE status = 'PENDING';
CREATE INDEX idx_outreach_messages_sending ON outreach_messages (updated_at) WHERE status = 'SENDING';
CREATE INDEX idx_outreach_messages_candidate ON outreach_messages (candidate_id, created_at DESC);
CREATE INDEX idx_outreach_messages_contact ON outreach_messages (candidate_id, contact_id);
CREATE INDEX idx_outreach_messages_account_sent ON outreach_messages (channel_account_id, sent_at)
    WHERE sent_at IS NOT NULL;
CREATE UNIQUE INDEX idx_outreach_messages_provider_id ON outreach_messages (provider_message_id)
    WHERE provider_message_id IS NOT NULL;
CREATE UNIQUE INDEX idx_outreach_messages_tracking ON outreach_messages (tracking_token)
    WHERE tracking_token IS NOT NULL;
