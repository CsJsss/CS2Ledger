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

func TestRunMatching_FIFO_SameAccount(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	buy := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-buy-1",
			ItemName:   "AK-47 | Redline",
			PaintSeed:  123,
			PaintIndex: 456,
			PaintWear:  0.07,
		},
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
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-sell-1",
			ItemName:   "AK-47 | Redline",
			PaintSeed:  123,
			PaintIndex: 456,
			PaintWear:  0.070001, // within ±0.0001 tolerance
		},
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
		AccountID: 1, CS2Item: model.CS2Item{AssetID: "asset-buy-1", ItemName: "AK-47 | Redline"},
		BuyTradeID: buy.ID, Status: "in_inventory",
	})

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 match, got %d", count)
	}

	sum, _ := svc.GetSummary(1)
	if sum == nil {
		t.Fatal("expected summary")
	}
	expectedNet := (1500-1000)*1 - (15 + 10)
	if sum.TotalNetPl != int64(expectedNet) {
		t.Errorf("expected net_pl %d, got %d", expectedNet, sum.TotalNetPl)
	}

	item, _ := ormInst.FindInventoryByAssetID(1, "asset-buy-1")
	if item != nil {
		t.Fatal("expected inventory item to be removed after sell match")
	}
}

func TestRunMatching_FIFO_CrossAccount(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	// Buy on BUFF (account 1).
	buy := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "buff-asset-1",
			ItemName:   "M4A1-S | Printstream",
			PaintSeed:  42,
			PaintIndex: 99,
			PaintWear:  0.03,
		},
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  2000,
		TotalPrice: 2000,
		Fee:        20,
		TradeAt:    1700000000000,
		ExternalID: "ext-buy-buff",
	}
	if err := ormInst.CreateTrade(buy); err != nil {
		t.Fatalf("create buy: %v", err)
	}

	_ = ormInst.UpsertInventory(&model.InventoryItem{
		AccountID: 1, CS2Item: model.CS2Item{AssetID: "buff-asset-1", ItemName: "M4A1-S | Printstream"},
		BuyTradeID: buy.ID, Status: "in_inventory",
	})

	// Sell on Youpin (account 2) — different asset_id, same item fingerprint.
	sell := &model.TradeRecord{
		AccountID: 2,
		CS2Item: model.CS2Item{
			AssetID:    "youpin-asset-99",
			ItemName:   "M4A1-S | Printstream",
			PaintSeed:  42,
			PaintIndex: 99,
			PaintWear:  0.03,
		},
		TradeType:  "sell",
		Quantity:   1,
		UnitPrice:  2500,
		TotalPrice: 2500,
		Fee:        25,
		TradeAt:    1700000000000,
		ExternalID: "ext-sell-youpin",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 cross-platform match, got %d", count)
	}

	// P&L attributed to sell account (account 2).
	sum, _ := svc.GetSummary(2)
	if sum == nil {
		t.Fatal("expected summary on sell account")
	}
	expectedNet := (2500-2000)*1 - (25 + 20)
	if sum.TotalNetPl != int64(expectedNet) {
		t.Errorf("expected net_pl %d, got %d", expectedNet, sum.TotalNetPl)
	}

	// Inventory removed from buy account (account 1).
	item, _ := ormInst.FindInventoryByAssetID(1, "buff-asset-1")
	if item != nil {
		t.Fatal("expected inventory removed from buy account after cross-platform sell")
	}
}

func TestRunMatching_NoMatch_DifferentFingerprint(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	buy := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-1",
			ItemName:   "AK-47 | Redline",
			PaintSeed:  111,
			PaintIndex: 222,
			PaintWear:  0.05,
		},
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  1000,
		TotalPrice: 1000,
		TradeAt:    1700000000000,
		ExternalID: "ext-buy-1",
	}
	if err := ormInst.CreateTrade(buy); err != nil {
		t.Fatalf("create buy: %v", err)
	}

	// Same item name but different paint_seed — different physical item.
	sell := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-2",
			ItemName:   "AK-47 | Redline",
			PaintSeed:  999,
			PaintIndex: 888,
			PaintWear:  0.01,
		},
		TradeType:  "sell",
		Quantity:   1,
		UnitPrice:  1500,
		TotalPrice: 1500,
		TradeAt:    1700000000000,
		ExternalID: "ext-sell-2",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 matches for different fingerprint, got %d", count)
	}
}

func TestRunMatching_FIFO_EarliestBuyFirst(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	// Buy 1: earliest (FIFO should pick this one).
	buy1 := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-early",
			ItemName:   "Desert Eagle | Blaze",
			PaintSeed:  1,
			PaintIndex: 1,
			PaintWear:  0.01,
		},
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  500,
		TotalPrice: 500,
		Fee:        5,
		TradeAt:    1600000000000, // earlier
		ExternalID: "ext-buy-early",
	}
	if err := ormInst.CreateTrade(buy1); err != nil {
		t.Fatalf("create buy1: %v", err)
	}

	// Buy 2: later, same item fingerprint.
	buy2 := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-late",
			ItemName:   "Desert Eagle | Blaze",
			PaintSeed:  1,
			PaintIndex: 1,
			PaintWear:  0.01,
		},
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  800,
		TotalPrice: 800,
		Fee:        8,
		TradeAt:    1700000000000, // later
		ExternalID: "ext-buy-late",
	}
	if err := ormInst.CreateTrade(buy2); err != nil {
		t.Fatalf("create buy2: %v", err)
	}

	_ = ormInst.UpsertInventory(&model.InventoryItem{
		AccountID: 1, CS2Item: model.CS2Item{AssetID: "asset-early", ItemName: "Desert Eagle | Blaze"},
		BuyTradeID: buy1.ID, Status: "in_inventory",
	})
	_ = ormInst.UpsertInventory(&model.InventoryItem{
		AccountID: 1, CS2Item: model.CS2Item{AssetID: "asset-late", ItemName: "Desert Eagle | Blaze"},
		BuyTradeID: buy2.ID, Status: "in_inventory",
	})

	sell := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID:    "asset-sell",
			ItemName:   "Desert Eagle | Blaze",
			PaintSeed:  1,
			PaintIndex: 1,
			PaintWear:  0.01,
		},
		TradeType:  "sell",
		Quantity:   1,
		UnitPrice:  1200,
		TotalPrice: 1200,
		TradeAt:    1800000000000,
		ExternalID: "ext-sell",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 match, got %d", count)
	}

	// P&L should use buy1 (the earliest, UnitPrice=500), not buy2 (UnitPrice=800).
	sum, _ := svc.GetSummary(1)
	expectedNet := (1200-500)*1 - (sell.Fee + buy1.Fee)
	if sum.TotalNetPl != expectedNet {
		t.Errorf("expected net_pl %d (FIFO picked earliest buy), got %d", expectedNet, sum.TotalNetPl)
	}

	// buy1's inventory should be removed (it was matched).
	item1, _ := ormInst.FindInventoryByAssetID(1, "asset-early")
	if item1 != nil {
		t.Fatal("expected early inventory removed")
	}

	// buy2's inventory should still exist (not matched yet).
	item2, _ := ormInst.FindInventoryByAssetID(1, "asset-late")
	if item2 == nil {
		t.Fatal("expected later inventory still present")
	}
}

func TestRunMatching_PartialFill_SellLargerThanBuy(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	// Buy qty=1, Sell qty=15 → only 1 unit matched, 14 unmatched.
	buy := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-buy", ItemName: "AWP | Dragon Lore",
			PaintSeed: 100, PaintIndex: 200, PaintWear: 0.01,
		},
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  1000,
		TotalPrice: 1000,
		TradeAt:    1700000000000,
		ExternalID: "ext-buy-qty1",
	}
	if err := ormInst.CreateTrade(buy); err != nil {
		t.Fatalf("create buy: %v", err)
	}

	sell := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-sell", ItemName: "AWP | Dragon Lore",
			PaintSeed: 100, PaintIndex: 200, PaintWear: 0.01,
		},
		TradeType:  "sell",
		Quantity:   15,
		UnitPrice:  1500,
		TotalPrice: 22500,
		TradeAt:    1800000000000,
		ExternalID: "ext-sell-qty15",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 match (only 1 buy unit available), got %d", count)
	}

	// P&L: only 1 unit matched.
	sum, _ := svc.GetSummary(1)
	expectedNet := (1500-1000)*1 - (sell.Fee*1/15 + buy.Fee)
	if sum.TotalNetPl != expectedNet {
		t.Errorf("expected net_pl %d, got %d", expectedNet, sum.TotalNetPl)
	}

	// Buy is fully consumed — inventory removed.
	item, _ := ormInst.FindInventoryByAssetID(1, "asset-buy")
	if item != nil {
		t.Fatal("expected buy inventory removed after full match")
	}

	// Sell has 14 unmatched units — should appear in unmatched sells.
	unmatched, _ := ormInst.FindUnmatchedSells(1)
	if len(unmatched) != 1 {
		t.Fatalf("expected 1 unmatched sell, got %d", len(unmatched))
	}
	if unmatched[0].ID != sell.ID {
		t.Fatal("expected the sell to be unmatched")
	}
}

func TestRunMatching_PartialFill_BuyLargerThanSell(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	// Buy qty=5, Sell qty=3 → sell fully matched, buy has 2 remaining.
	buy := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-buy", ItemName: "AK-47 | Redline",
			PaintSeed: 1, PaintIndex: 1, PaintWear: 0.05,
		},
		TradeType:  "buy",
		Quantity:   5,
		UnitPrice:  1000,
		TotalPrice: 5000,
		Fee:        10,
		TradeAt:    1700000000000,
		ExternalID: "ext-buy-qty5",
	}
	if err := ormInst.CreateTrade(buy); err != nil {
		t.Fatalf("create buy: %v", err)
	}

	sell := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-sell", ItemName: "AK-47 | Redline",
			PaintSeed: 1, PaintIndex: 1, PaintWear: 0.05,
		},
		TradeType:  "sell",
		Quantity:   3,
		UnitPrice:  1500,
		TotalPrice: 4500,
		Fee:        15,
		TradeAt:    1800000000000,
		ExternalID: "ext-sell-qty3",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 match, got %d", count)
	}

	// P&L: 3 units matched.
	sum, _ := svc.GetSummary(1)
	expectedNet := (1500-1000)*3 - (sell.Fee*3/3 + buy.Fee*3/5)
	if sum.TotalNetPl != expectedNet {
		t.Errorf("expected net_pl %d, got %d", expectedNet, sum.TotalNetPl)
	}

	// Buy has 2 remaining units — inventory should still exist with qty=2.
	item, _ := ormInst.FindInventoryByAssetID(1, "asset-buy")
	if item == nil {
		t.Fatal("expected inventory still present with remaining quantity")
	}
	if item.Quantity != 2 {
		t.Fatalf("expected inventory qty 2, got %d", item.Quantity)
	}

	// Sell is fully matched — no unmatched sells.
	unmatched, _ := ormInst.FindUnmatchedSells(1)
	if len(unmatched) != 0 {
		t.Fatalf("expected 0 unmatched sells, got %d", len(unmatched))
	}
}

func TestRunMatching_PartialFill_MultipleBuys(t *testing.T) {
	svc, ormInst, cleanup := setupService(t)
	defer cleanup()

	// Buy1 qty=5 (earlier), Buy2 qty=3 (later), Sell qty=7.
	// FIFO: 5 from buy1, 2 from buy2.
	buy1 := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-buy1", ItemName: "M4A4 | Howl",
			PaintSeed: 5, PaintIndex: 5, PaintWear: 0.02,
		},
		TradeType:  "buy",
		Quantity:   5,
		UnitPrice:  2000,
		TotalPrice: 10000,
		Fee:        20,
		TradeAt:    1600000000000,
		ExternalID: "ext-buy1-qty5",
	}
	if err := ormInst.CreateTrade(buy1); err != nil {
		t.Fatalf("create buy1: %v", err)
	}

	buy2 := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-buy2", ItemName: "M4A4 | Howl",
			PaintSeed: 5, PaintIndex: 5, PaintWear: 0.02,
		},
		TradeType:  "buy",
		Quantity:   3,
		UnitPrice:  2500,
		TotalPrice: 7500,
		Fee:        25,
		TradeAt:    1700000000000,
		ExternalID: "ext-buy2-qty3",
	}
	if err := ormInst.CreateTrade(buy2); err != nil {
		t.Fatalf("create buy2: %v", err)
	}

	sell := &model.TradeRecord{
		AccountID: 1,
		CS2Item: model.CS2Item{
			AssetID: "asset-sell", ItemName: "M4A4 | Howl",
			PaintSeed: 5, PaintIndex: 5, PaintWear: 0.02,
		},
		TradeType:  "sell",
		Quantity:   7,
		UnitPrice:  3000,
		TotalPrice: 21000,
		Fee:        30,
		TradeAt:    1800000000000,
		ExternalID: "ext-sell-qty7",
	}
	if err := ormInst.CreateTrade(sell); err != nil {
		t.Fatalf("create sell: %v", err)
	}

	count, err := svc.RunMatching()
	if err != nil {
		t.Fatalf("RunMatching: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 matches (2 buys consumed), got %d", count)
	}

	// buy1 fully consumed (qty=5 matched) → inventory removed.
	item1, _ := ormInst.FindInventoryByAssetID(1, "asset-buy1")
	if item1 != nil {
		t.Fatal("expected buy1 inventory removed")
	}

	// buy2 partially consumed (qty=2 matched, 1 remaining).
	item2, _ := ormInst.FindInventoryByAssetID(1, "asset-buy2")
	if item2 == nil {
		t.Fatal("expected buy2 inventory present")
	}
	if item2.Quantity != 1 {
		t.Fatalf("expected buy2 inventory qty 1, got %d", item2.Quantity)
	}

	// P&L: 5 units from buy1 at 2000, 2 units from buy2 at 2500, sell at 3000.
	sum, _ := svc.GetSummary(1)
	expectedGross := (3000-2000)*5 + (3000-2500)*2 // 5000 + 1000 = 6000
	if sum.TotalGrossPl != int64(expectedGross) {
		t.Errorf("expected gross_pl %d, got %d", expectedGross, sum.TotalGrossPl)
	}

	// Sell fully matched — no unmatched sells.
	unmatched, _ := ormInst.FindUnmatchedSells(1)
	if len(unmatched) != 0 {
		t.Fatalf("expected 0 unmatched sells, got %d", len(unmatched))
	}
}
