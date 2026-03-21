CREATE TABLE invitations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email       VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'member',
    invited_by  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_invitations_token UNIQUE (token),
    CONSTRAINT chk_invitations_role CHECK (role IN ('owner', 'admin', 'member', 'viewer'))
);

CREATE INDEX idx_invitations_org_id ON invitations (org_id);
CREATE INDEX idx_invitations_email ON invitations (email);
CREATE INDEX idx_invitations_token ON invitations (token);
