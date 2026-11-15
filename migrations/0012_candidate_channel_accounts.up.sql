CREATE TABLE IF NOT EXISTS candidate_channel_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    channel_type VARCHAR(20) NOT NULL CHECK (channel_type IN ('EMAIL')),
    identifier VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    smtp_host VARCHAR(255) NOT NULL,
    smtp_port INTEGER NOT NULL CHECK (smtp_port BETWEEN 1 AND 65535),
    smtp_security VARCHAR(10) NOT NULL CHECK (smtp_security IN ('TLS', 'STARTTLS', 'NONE')),
    imap_host VARCHAR(255) NOT NULL,
    imap_port INTEGER NOT NULL CHECK (imap_port BETWEEN 1 AND 65535),
    imap_security VARCHAR(10) NOT NULL CHECK (imap_security IN ('TLS', 'STARTTLS', 'NONE')),
    encrypted_credentials BYTEA NOT NULL,
    key_version INTEGER NOT NULL CHECK (key_version > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'UNVERIFIED' CHECK (status IN ('UNVERIFIED', 'ACTIVE', 'FAILED')),
    daily_limit INTEGER NOT NULL CHECK (daily_limit > 0),
    last_verified_at TIMESTAMPTZ,
    last_error TEXT,
    imap_uid_validity BIGINT,
    imap_last_uid BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_channel_accounts_identifier
    ON candidate_channel_accounts (candidate_id, channel_type, LOWER(identifier));
CREATE INDEX idx_channel_accounts_active ON candidate_channel_accounts (id) WHERE status = 'ACTIVE';
