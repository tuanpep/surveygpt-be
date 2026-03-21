CREATE TABLE templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID REFERENCES organizations(id) ON DELETE CASCADE,
    category        VARCHAR(100) NOT NULL DEFAULT 'general',
    title           VARCHAR(500) NOT NULL,
    description     TEXT DEFAULT '',
    tags            TEXT[] NOT NULL DEFAULT '{}',
    structure       JSONB NOT NULL DEFAULT '{}',
    theme           JSONB NOT NULL DEFAULT '{}',
    cover_image_url TEXT,
    is_featured     BOOLEAN NOT NULL DEFAULT FALSE,
    use_count       INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_templates_org_id ON templates (org_id);
CREATE INDEX idx_templates_category ON templates (category);
CREATE INDEX idx_templates_is_featured ON templates (is_featured);
CREATE INDEX idx_templates_use_count ON templates (use_count DESC);
CREATE INDEX idx_templates_tags ON templates USING gin (tags);
