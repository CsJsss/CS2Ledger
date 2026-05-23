package orm

import (
	"testing"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

func setupTestDB(t *testing.T) (ORMInterface, func()) {
	t.Helper()

	orm, _, err := NewTestORM()
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	cleanup := func() {
		// ormImpl holds the connection internally
	}

	return orm, cleanup
}

func TestAccountCRUD(t *testing.T) {
	orm, cleanup := setupTestDB(t)
	defer cleanup()

	a := &model.Account{
		Name:     "test-account",
		Platform: "buff",
		Cookie:   "test-cookie",
		Status:   AccountStatusActive,
	}
	if err := orm.CreateAccount(a); err != nil {
		t.Fatalf("create: %v", err)
	}
	if a.ID == 0 {
		t.Fatal("expected ID to be set")
	}

	all, err := orm.ListAccounts()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 account, got %d", len(all))
	}

	found, err := orm.FindAccountByID(a.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if found.Name != "test-account" {
		t.Fatalf("expected name test-account, got %s", found.Name)
	}

	a.Name = "updated"
	if err := orm.UpdateAccount(a); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := orm.DeleteAccount(a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	all, _ = orm.ListAccounts()
	if len(all) != 0 {
		t.Fatalf("expected 0 accounts after delete, got %d", len(all))
	}
}

func TestInventoryUpsert(t *testing.T) {
	orm, cleanup := setupTestDB(t)
	defer cleanup()

	item := &model.InventoryItem{
		AccountID:  1,
		CS2Item:    model.CS2Item{AssetID: "asset-1", ItemName: "AK-47 | Redline"},
		BuyTradeID: 1,
		Status:     "in_inventory",
	}
	if err := orm.UpsertInventory(item); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	items, err := orm.FindInventoryByAccount(1, "")
	if err != nil {
		t.Fatalf("find by account: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item.ItemName = "Updated Name"
	item.Status = "listed"
	if err := orm.UpsertInventory(item); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	items, _ = orm.FindInventoryByAccount(1, "")
	if len(items) != 1 {
		t.Fatalf("expected 1 item after update, got %d", len(items))
	}
	if items[0].Status != "listed" {
		t.Fatalf("expected status listed, got %s", items[0].Status)
	}

	if err := orm.RemoveInventoryByAssetID(1, "asset-1"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	items, _ = orm.FindInventoryByAccount(1, "")
	if len(items) != 0 {
		t.Fatalf("expected 0 items after remove, got %d", len(items))
	}
}

func TestPnlUpsertDaily(t *testing.T) {
	orm, cleanup := setupTestDB(t)
	defer cleanup()

	tradeAt := int64(1700000000000)

	if err := orm.UpsertDailyPnl(1, tradeAt, 1000, 100, 900); err != nil {
		t.Fatalf("upsert daily: %v", err)
	}

	if err := orm.UpsertDailyPnl(1, tradeAt, 500, 50, 450); err != nil {
		t.Fatalf("upsert daily 2: %v", err)
	}

	records, err := orm.FindPnlByAccount(1)
	if err != nil {
		t.Fatalf("find by account: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 daily record, got %d", len(records))
	}

	r := records[0]
	if r.TradeCount != 2 {
		t.Fatalf("expected trade_count 2, got %d", r.TradeCount)
	}
	if r.GrossPl != 1500 {
		t.Fatalf("expected gross_pl 1500, got %d", r.GrossPl)
	}
	if r.NetPl != 1350 {
		t.Fatalf("expected net_pl 1350, got %d", r.NetPl)
	}
}
