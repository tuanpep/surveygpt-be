CREATE TABLE email_lists (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT DEFAULT '',
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    contact_count   INTEGER NOT NULL DEFAULT 0,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_lists_org_id ON email_lists (org_id);

CREATE TABLE email_contacts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_list_id   UUID NOT NULL REFERENCES email_lists(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    first_name      VARCHAR(255) DEFAULT '',
    last_name       VARCHAR(255) DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}',
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_email_contacts_list_email UNIQUE (email_list_id, email)
);

CREATE INDEX idx_email_contacts_email_list_id ON email_contacts (email_list_id);
CREATE INDEX idx_email_contacts_email ON email_contacts (email);
CREATE INDEX idx_email_contacts_status ON email_contacts (status);
