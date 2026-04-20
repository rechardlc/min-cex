CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id    BIGINT      NOT NULL,
    balance    NUMERIC(20, 8) NOT NULL DEFAULT 0,
    CONSTRAINT uq_users_user_id UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
