CREATE TABLE surveys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title           VARCHAR(500) NOT NULL DEFAULT '',
    description     TEXT DEFAULT '',
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',
    ui_mode         VARCHAR(50) NOT NULL DEFAULT 'classic',
    structure       JSONB NOT NULL DEFAULT '{}',
    settings        JSONB NOT NULL DEFAULT '{}',
    theme           JSONB NOT NULL DEFAULT '{}',
    response_count  INTEGER NOT NULL DEFAULT 0,
    view_count      INTEGER NOT NULL DEFAULT 0,
    published_at    TIMESTAMPTZ,
    closed_at       TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_surveys_status CHECK (status IN ('draft', 'published', 'closed', 'archived')),
    CONSTRAINT chk_surveys_ui_mode CHECK (ui_mode IN ('classic', 'minimal', 'cards', 'conversational'))
);

CREATE INDEX idx_surveys_org_id ON surveys (org_id);
CREATE INDEX idx_surveys_created_by ON surveys (created_by);
CREATE INDEX idx_surveys_status ON surveys (status);
CREATE INDEX idx_surveys_created_at ON surveys (created_at);
CREATE INDEX idx_surveys_deleted_at ON surveys (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_surveys_structure ON surveys USING gin (structure);
