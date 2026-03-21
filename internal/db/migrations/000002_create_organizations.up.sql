CREATE TABLE organizations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                    VARCHAR(255) NOT NULL,
    slug                    VARCHAR(100) NOT NULL,
    plan                    VARCHAR(50) NOT NULL DEFAULT 'free',
    stripe_customer_id      VARCHAR(255),
    stripe_subscription_id  VARCHAR(255),
    billing_email           VARCHAR(255),
    settings                JSONB NOT NULL DEFAULT '{}',
    response_limit          INTEGER NOT NULL DEFAULT 100,
    survey_limit            INTEGER NOT NULL DEFAULT 10,
    member_limit            INTEGER NOT NULL DEFAULT 5,
    ai_credits              INTEGER NOT NULL DEFAULT 50,
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_organizations_slug ON organizations (slug);
CREATE INDEX idx_organizations_slug ON organizations (slug);
CREATE INDEX idx_organizations_plan ON organizations (plan);
CREATE INDEX idx_organizations_stripe_customer_id ON organizations (stripe_customer_id);
