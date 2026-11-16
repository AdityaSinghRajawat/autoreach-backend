CREATE TABLE IF NOT EXISTS outreach_campaign_steps (
    campaign_id UUID NOT NULL REFERENCES outreach_campaigns(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL CHECK (step_number > 0),
    template_id UUID NOT NULL REFERENCES outreach_templates(id) ON DELETE RESTRICT,
    delay_days INTEGER NOT NULL CHECK (delay_days >= 0),
    PRIMARY KEY (campaign_id, step_number)
);

CREATE INDEX idx_outreach_campaign_steps_template ON outreach_campaign_steps (template_id);
