CREATE TABLE answers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    response_id UUID NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
    question_id VARCHAR(36) NOT NULL,
    value       JSONB NOT NULL DEFAULT '{}',
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_answers_response_question UNIQUE (response_id, question_id)
);

CREATE INDEX idx_answers_response_id ON answers (response_id);
CREATE INDEX idx_answers_question_id ON answers (question_id);
CREATE INDEX idx_answers_value ON answers USING gin (value);
