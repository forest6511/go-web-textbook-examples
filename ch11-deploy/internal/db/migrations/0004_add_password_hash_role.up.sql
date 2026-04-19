ALTER TABLE users
    ADD COLUMN password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN role TEXT NOT NULL DEFAULT 'user'
        CHECK (role IN ('user', 'admin'));

ALTER TABLE users ALTER COLUMN password_hash DROP DEFAULT;

-- Ch 07 では password_hash を使うため、旧 password カラムは不要
ALTER TABLE users DROP COLUMN password;
