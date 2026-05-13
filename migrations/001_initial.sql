CREATE TABLE accounts (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    name              TEXT NOT NULL,
    platform          TEXT NOT NULL,
    cookie            TEXT NOT NULL,
    available_balance INTEGER NOT NULL DEFAULT 0,
    purchase_balance  INTEGER NOT NULL DEFAULT 0,
    remark            TEXT DEFAULT '',
    status            TEXT NOT NULL DEFAULT 'active',
    last_sync_at      INTEGER,
    created_at        DATETIME NOT NULL,
    updated_at        DATETIME NOT NULL,
    deleted_at        DATETIME
);

CREATE UNIQUE INDEX idx_accounts_name ON accounts(name);

CREATE TABLE trade_records (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id           INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    trade_type           TEXT NOT NULL,
    quantity             INTEGER NOT NULL DEFAULT 1,
    consumed_quantity    INTEGER NOT NULL DEFAULT 0,
    unit_price           INTEGER NOT NULL,
    total_price          INTEGER NOT NULL,
    fee                  INTEGER NOT NULL DEFAULT 0,
    trade_at             INTEGER NOT NULL,
    source               TEXT NOT NULL DEFAULT '',
    state                TEXT NOT NULL DEFAULT 'SUCCESS',
    state_text           TEXT DEFAULT '',
    transact_time        INTEGER,
    trade_offer_id       TEXT DEFAULT '',
    external_id          TEXT,
    matched_buy_trade_id INTEGER REFERENCES trade_records(id),
    remark               TEXT DEFAULT '',

    -- CS2Item 内嵌字段
    category_group          TEXT DEFAULT '',
    asset_id                 TEXT NOT NULL,
    class_id                 TEXT DEFAULT '',
    instance_id              TEXT DEFAULT '',
    goods_id                 INTEGER NOT NULL DEFAULT 0,
    market_hash_name         TEXT DEFAULT '',
    item_name                TEXT NOT NULL,
    icon_url                 TEXT DEFAULT '',
    paint_wear               REAL NOT NULL DEFAULT 0,
    paint_seed               INTEGER NOT NULL DEFAULT 0,
    paint_index              INTEGER NOT NULL DEFAULT 0,
    exterior                 TEXT DEFAULT '',
    rarity                   TEXT DEFAULT '',
    weapon_type              TEXT DEFAULT '',
    weapon_name              TEXT DEFAULT '',
    quality                  TEXT DEFAULT '',
    series                   TEXT DEFAULT '',
    itemset                  TEXT DEFAULT '',
    weapon_case              TEXT DEFAULT '',
    custom                   TEXT DEFAULT '',
    stickers                 TEXT,
    keychains                TEXT,
    tradable_unfrozen_time   INTEGER,

    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME
);

CREATE UNIQUE INDEX idx_trades_external ON trade_records(account_id, external_id)
    WHERE external_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_trades_account ON trade_records(account_id, trade_at DESC);
CREATE INDEX idx_trades_asset ON trade_records(asset_id);
CREATE INDEX idx_trades_type ON trade_records(account_id, trade_type);
CREATE INDEX idx_trades_matched ON trade_records(matched_buy_trade_id)
    WHERE matched_buy_trade_id IS NOT NULL;
CREATE INDEX idx_trades_type_at ON trade_records(trade_type, trade_at);

CREATE TABLE inventory (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id   INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    buy_trade_id INTEGER NOT NULL REFERENCES trade_records(id),
    status       TEXT NOT NULL DEFAULT 'in_inventory',
    listed_price INTEGER,
    listed_at    INTEGER,
    quantity     INTEGER NOT NULL DEFAULT 1,

    -- CS2Item 内嵌字段
    category_group          TEXT DEFAULT '',
    asset_id                 TEXT NOT NULL,
    class_id                 TEXT DEFAULT '',
    instance_id              TEXT DEFAULT '',
    goods_id                 INTEGER NOT NULL DEFAULT 0,
    market_hash_name         TEXT DEFAULT '',
    item_name                TEXT NOT NULL,
    icon_url                 TEXT DEFAULT '',
    paint_wear               REAL NOT NULL DEFAULT 0,
    paint_seed               INTEGER NOT NULL DEFAULT 0,
    paint_index              INTEGER NOT NULL DEFAULT 0,
    exterior                 TEXT DEFAULT '',
    rarity                   TEXT DEFAULT '',
    weapon_type              TEXT DEFAULT '',
    weapon_name              TEXT DEFAULT '',
    quality                  TEXT DEFAULT '',
    series                   TEXT DEFAULT '',
    itemset                  TEXT DEFAULT '',
    weapon_case              TEXT DEFAULT '',
    custom                   TEXT DEFAULT '',
    stickers                 TEXT,
    keychains                TEXT,
    tradable_unfrozen_time   INTEGER,

    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME
);

CREATE UNIQUE INDEX idx_inventory_asset ON inventory(account_id, asset_id);
CREATE INDEX idx_inventory_status ON inventory(account_id, status);

CREATE TABLE rental_records (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id    INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    asset_id      TEXT NOT NULL,
    item_name     TEXT NOT NULL,
    income        INTEGER NOT NULL,
    duration_days INTEGER NOT NULL,
    start_at      INTEGER NOT NULL,
    end_at        INTEGER NOT NULL,
    external_id   TEXT,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    deleted_at    DATETIME
);

CREATE INDEX idx_rental_account ON rental_records(account_id, start_at DESC);
CREATE INDEX idx_rental_asset ON rental_records(asset_id);
CREATE UNIQUE INDEX idx_rental_external ON rental_records(account_id, external_id)
    WHERE external_id IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE pnl_daily (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id  INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    date        TEXT NOT NULL,
    trade_count INTEGER NOT NULL DEFAULT 0,
    gross_pl    INTEGER NOT NULL DEFAULT 0,
    fee         INTEGER NOT NULL DEFAULT 0,
    net_pl      INTEGER NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    deleted_at  DATETIME,
    UNIQUE(account_id, date)
);

CREATE INDEX idx_pnl_daily_account ON pnl_daily(account_id, date DESC);

CREATE TABLE schema_version (
    version    INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL
);
