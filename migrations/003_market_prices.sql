CREATE TABLE IF NOT EXISTS market_prices (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    market_hash_name TEXT NOT NULL UNIQUE,
    buff_price       REAL,
    buff_volume      INTEGER,
    youpin_price     REAL,
    youpin_volume    INTEGER,
    steam_price      REAL,
    steam_volume     INTEGER,
    updated_at       INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_market_prices_updated ON market_prices(updated_at);

ALTER TABLE trade_records ADD COLUMN csqaq_goods_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE inventory ADD COLUMN csqaq_goods_id INTEGER NOT NULL DEFAULT 0;
