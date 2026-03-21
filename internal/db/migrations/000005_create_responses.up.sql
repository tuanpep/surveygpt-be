CREATE TABLE responses (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id           UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    respondent_email    VARCHAR(255),
    respondent_name     VARCHAR(255),
    respondent_device   VARCHAR(100),
    respondent_browser  VARCHAR(100),
    respondent_os       VARCHAR(100),
    respondent_country  VARCHAR(100),
    respondent_ip_hash  VARCHAR(64),
    status              VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    duration_ms         INTEGER,
    source              VARCHAR(50) NOT NULL DEFAULT 'direct',
    source_detail       VARCHAR(255),
    language            VARCHAR(10) NOT NULL DEFAULT 'en',
    embedded_data       JSONB NOT NULL DEFAULT '{}',
    quality_flags       JSONB NOT NULL DEFAULT '{}',
    ai_analysis         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_responses_status CHECK (status IN ('in_progress', 'completed', 'disqualified', 'partial'))
);

CREATE INDEX idx_responses_survey_id ON responses (survey_id);
CREATE INDEX idx_responses_status ON responses (status);
CREATE INDEX idx_responses_created_at ON responses (created_at);
CREATE INDEX idx_responses_completed_at ON responses (completed_at) WHERE completed_at IS NOT NULL;
CREATE INDEX idx_responses_respondent_email ON responses (respondent_email);
CREATE INDEX idx_responses_survey_created ON responses (survey_id, created_at DESC);
