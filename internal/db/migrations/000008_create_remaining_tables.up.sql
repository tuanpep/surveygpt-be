-- AI Credits
CREATE TABLE ai_credits (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    total_credits  INTEGER NOT NULL DEFAULT 50,
    used_credits   INTEGER NOT NULL DEFAULT 0,
    period_start   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period_end     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_ai_credits_org_id UNIQUE (org_id)
);

CREATE INDEX idx_ai_credits_org_id ON ai_credits (org_id);

-- Analytics Cache
CREATE TABLE analytics_cache (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id    UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    cache_type   VARCHAR(100) NOT NULL,
    data         JSONB NOT NULL DEFAULT '{}',
    period_start TIMESTAMPTZ,
    period_end   TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_analytics_cache_survey_type UNIQUE (survey_id, cache_type)
);

CREATE INDEX idx_analytics_cache_survey_id ON analytics_cache (survey_id);
CREATE INDEX idx_analytics_cache_cache_type ON analytics_cache (cache_type);
CREATE INDEX idx_analytics_cache_expires_at ON analytics_cache (expires_at);

-- API Keys
CREATE TABLE api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    key_hash    VARCHAR(255) NOT NULL,
    key_prefix  VARCHAR(16) NOT NULL,
    scopes      TEXT[] NOT NULL DEFAULT '{}',
    last_used   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at  TIMESTAMPTZ,

    CONSTRAINT uq_api_keys_key_hash UNIQUE (key_hash)
);

CREATE INDEX idx_api_keys_org_id ON api_keys (org_id);
CREATE INDEX idx_api_keys_user_id ON api_keys (user_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys (key_prefix);

-- Audit Logs
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID REFERENCES organizations(id) ON DELETE SET NULL,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    action      VARCHAR(100) NOT NULL,
    resource    VARCHAR(100) NOT NULL,
    resource_id UUID,
    details     JSONB NOT NULL DEFAULT '{}',
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org_id ON audit_logs (org_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs (action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at);

-- Files
CREATE TABLE files (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    uploaded_by   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    filename      VARCHAR(500) NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    mime_type     VARCHAR(255) NOT NULL,
    size_bytes    BIGINT NOT NULL DEFAULT 0,
    storage_key   VARCHAR(1000) NOT NULL,
    public_url    TEXT,
    purpose       VARCHAR(100) NOT NULL DEFAULT 'general',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_org_id ON files (org_id);
CREATE INDEX idx_files_uploaded_by ON files (uploaded_by);
CREATE INDEX idx_files_purpose ON files (purpose);
CREATE INDEX idx_files_storage_key ON files (storage_key);

-- Integrations
CREATE TABLE integrations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider        VARCHAR(100) NOT NULL,
    config          JSONB NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_integrations_org_provider UNIQUE (org_id, provider)
);

CREATE INDEX idx_integrations_org_id ON integrations (org_id);
CREATE INDEX idx_integrations_provider ON integrations (provider);

-- Usage Logs
CREATE TABLE usage_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    action      VARCHAR(100) NOT NULL,
    resource    VARCHAR(100) NOT NULL,
    quantity    INTEGER NOT NULL DEFAULT 1,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_logs_org_id ON usage_logs (org_id);
CREATE INDEX idx_usage_logs_action ON usage_logs (action);
CREATE INDEX idx_usage_logs_created_at ON usage_logs (created_at);
CREATE INDEX idx_usage_logs_org_created ON usage_logs (org_id, created_at DESC);

-- Webhooks
CREATE TABLE webhooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    url             TEXT NOT NULL,
    events          TEXT[] NOT NULL DEFAULT '{}',
    secret          VARCHAR(255) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_triggered  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_org_id ON webhooks (org_id);
