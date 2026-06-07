CREATE TABLE IF NOT EXISTS bill_records (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id  INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    platform    TEXT NOT NULL,
    type_id     INTEGER NOT NULL,
    type_name   TEXT NOT NULL,
    this_money  INTEGER NOT NULL,
    order_no    TEXT DEFAULT '',
    add_time    INTEGER NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    deleted_at  DATETIME
);

CREATE INDEX IF NOT EXISTS idx_bill_account ON bill_records(account_id, add_time DESC);
CREATE INDEX IF NOT EXISTS idx_bill_order ON bill_records(order_no);

-- Speed up filtered pagination (account_id + type_id + time sort)
-- Covers: WHERE account_id=? AND type_id=? ORDER BY add_time DESC
-- Also helps: SumRentalIncome WHERE account_id=? AND type_id IN (...)
CREATE INDEX IF NOT EXISTS idx_bill_account_type_time ON bill_records(account_id, type_id, add_time DESC);

-- Speed up platform filter combined with account
-- Covers: WHERE account_id=? AND platform=? ORDER BY add_time DESC
CREATE INDEX IF NOT EXISTS idx_bill_account_platform_time ON bill_records(account_id, platform, add_time DESC);


ALTER TABLE accounts ADD COLUMN bill_last_sync_at INTEGER;