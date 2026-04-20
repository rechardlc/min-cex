CREATE TABLE IF NOT EXISTS trades (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    market     VARCHAR(20)    NOT NULL,
    price      NUMERIC(20, 8) NOT NULL,
    size       NUMERIC(20, 8) NOT NULL,
    bid        BOOLEAN        NOT NULL,
    bid_user_id BIGINT        NOT NULL,
    ask_user_id BIGINT        NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_trades_deleted_at ON trades (deleted_at);
CREATE INDEX IF NOT EXISTS idx_trades_market     ON trades (market);
CREATE INDEX IF NOT EXISTS idx_trades_created_at ON trades (created_at DESC);
