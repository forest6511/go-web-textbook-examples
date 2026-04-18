CREATE TABLE audits (
    id         BIGSERIAL PRIMARY KEY,
    action     TEXT NOT NULL,
    task_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
