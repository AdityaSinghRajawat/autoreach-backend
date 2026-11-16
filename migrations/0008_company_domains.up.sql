CREATE TABLE IF NOT EXISTS company_domains (
    normalized_name VARCHAR(255) PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    company_domain VARCHAR(255),
    provider VARCHAR(30),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
