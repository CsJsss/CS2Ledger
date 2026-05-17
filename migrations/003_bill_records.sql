CREATE TABLE bill_records (
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

CREATE INDEX idx_bill_account ON bill_records(account_id, add_time DESC);
CREATE INDEX idx_bill_order ON bill_records(order_no);
