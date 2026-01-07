-- migration: 20260106_create_cost_records.up.sql
CREATE TABLE IF NOT EXISTS cost_records (
    id SERIAL PRIMARY KEY,
    sn VARCHAR(64) NOT NULL,
    model_name VARCHAR(64) NOT NULL,
    prompt_tokens INTEGER NOT NULL,
    completion_tokens INTEGER NOT NULL,
    cost NUMERIC(10,6) NOT NULL,
    currency CHAR(3) DEFAULT 'CNY',
    duration_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_cost_sn ON cost_records(sn);
CREATE INDEX IF NOT EXISTS idx_cost_model_name ON cost_records(model_name);
CREATE INDEX IF NOT EXISTS idx_cost_created_at ON cost_records(created_at);
CREATE INDEX IF NOT EXISTS idx_cost_deleted_at ON cost_records(deleted_at) WHERE deleted_at IS NULL;
