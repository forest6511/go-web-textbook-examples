CREATE TABLE attachments (
    id              UUID PRIMARY KEY,
    owner_id        BIGINT NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    object_key      TEXT NOT NULL UNIQUE,
    filename        TEXT NOT NULL,
    content_type    TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL,
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_owner_id
    ON attachments(owner_id);
