CREATE TABLE org_memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(50) NOT NULL DEFAULT 'member',
    invited_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_org_memberships_org_user UNIQUE (org_id, user_id),
    CONSTRAINT chk_org_memberships_role CHECK (role IN ('owner', 'admin', 'member', 'viewer'))
);

CREATE INDEX idx_org_memberships_org_id ON org_memberships (org_id);
CREATE INDEX idx_org_memberships_user_id ON org_memberships (user_id);
CREATE INDEX idx_org_memberships_role ON org_memberships (role);
