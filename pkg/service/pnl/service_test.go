package pnl

import (
	"testing"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
)

func setupService(t *testing.T) (PnlInterface, orm.ORMInterface, func()) {
	t.Helper()

	ormInst, err := orm.NewTestORM()
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	cleanup := func() {}

	return NewService(nil, ormInst), ormInst, cleanup
}

func TestProcessPending_SingleMatch(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	buy := &model.TradeRecord{
		AccountID:  1,
		AssetID:    "asset-1",
		ItemName:   "AK-47 | Redline",
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  1000,
		TotalPrice: 1000,
		Fee:        10,
		TradeAt:    1700000000000,
		ExternalID: "ext-buy-1",
	}
	if err := ormInst.CreateTrade(buy); err != nil {
		t.Fatalf("create buy: %v", err)
	}

	sell := &model.TradeRecord{
		AccountID:  1,
		AssetID:    "asset-1",
		ItemName:   "AK-47 | Redline",
		TradeType:  "sell",
		Quantity:   1,
		UnitPrice:  1500,
		TotalPrice: 1500,
		Fee:        15,
		TradeAt:    1700000000000,
		ExternalID: "ext-sell-1",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	_ = ormInst.UpsertInventory(&model.InventoryItem{
		AccountID: 1, AssetID: "asset-1", ItemName: "AK-47 | Redline",
		BuyTradeID: buy.ID, Status: "in_inventory",
	})

	count, err := svc.ProcessPending(1)
	if err != nil {
		t.Fatalf("ProcessPending: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 match, got %d", count)
	}

	sum, _ := svc.GetSummary(1)
	if sum == nil {
		t.Fatal("expected summary")
	}
	if sum.TotalTrades != 1 {
		t.Errorf("expected 1 trade in pnl, got %d", sum.TotalTrades)
	}
	expectedNet := (1500-1000)*1 - (15 + 10)
	if sum.TotalNetPl != int64(expectedNet) {
		t.Errorf("expected net_pl %d, got %d", expectedNet, sum.TotalNetPl)
	}

	item, _ := ormInst.FindInventoryByAssetID(1, "asset-1")
	if item != nil {
		t.Fatal("expected inventory item to be removed after sell match")
	}
}
