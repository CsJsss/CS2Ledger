package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/factory"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/utils"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

// isBulkItem returns true for fungible items without unique paint attributes
// (cases, stickers, keys, capsules, etc.). These are bought/sold in volume and
// would otherwise generate excessive individual trade records.
func isBulkItem(item model.CS2Item) bool {
	return item.Exterior == "" && item.PaintSeed == 0 && item.PaintIndex == 0 && item.PaintWear == 0
}

type Engine struct {
	log     *logfx.Logger
	factory *factory.PlatformFactory
	orm     orm.ORMInterface
	pnlSvc  pnl.PnlInterface
	mu      sync.Mutex
}

func NewEngine(
	log *logfx.Logger,
	f *factory.PlatformFactory,
	orm orm.ORMInterface,
	pnlSvc pnl.PnlInterface,
) *Engine {
	return &Engine{
		log:     log,
		factory: f,
		orm:     orm,
		pnlSvc:  pnlSvc,
	}
}

type SyncResult struct {
	NewTrades int
	NewPnl    int
	Warnings  []string `json:"warnings"`
}

func (e *Engine) SyncAccount(accountID uint) (*SyncResult, error) {
	acc, err := e.loadAccount(accountID)
	if err != nil {
		return nil, err
	}

	e.log.Info("sync started", "account_id", accountID, "platform", acc.Platform, "name", acc.Name)

	client, err := e.createAndVerify(acc)
	if err != nil {
		return nil, err
	}

	buys, sells, balance, warnings := e.fetchData(client, acc.LastSyncAt)
	result := &SyncResult{Warnings: warnings}

	maxTradeMs := maxTradeAt(buys, sells)

	buys = aggregateBulkTrades(buys, acc.Platform, model.DirectionBuy)
	sells = aggregateBulkTrades(sells, acc.Platform, model.DirectionSell)

	e.log.Debug("saving trades", "buys", len(buys), "sells", len(sells))

	result.NewTrades = e.persistTrades(accountID, acc.Platform, buys, sells)

	e.mu.Lock()
	matchCount, matchErr := e.pnlSvc.RunMatching()
	e.mu.Unlock()
	result.NewPnl = matchCount
	e.log.Debug("global matching completed", "matched", matchCount, "err", matchErr)

	e.updateAccountMeta(accountID, balance, maxTradeMs, len(warnings) > 0)

	e.log.Info("completed",
		"new_trades", result.NewTrades,
		"new_pnl", result.NewPnl,
		"warnings", len(warnings),
	)

	return result, nil
}

// loadAccount retrieves the account or returns an error.
func (e *Engine) loadAccount(accountID uint) (*model.Account, error) {
	acc, err := e.orm.FindAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return acc, nil
}

// createAndVerify builds the platform client, verifies credentials, and
// updates the account status accordingly.
func (e *Engine) createAndVerify(acc *model.Account) (platform.Client, error) {
	client, err := e.factory.New(acc.Platform, acc.Cookie, e.log)
	if err != nil {
		e.log.Error("create client failed", "err", err)
		return nil, fmt.Errorf("create client: %w", err)
	}

	ctx := context.Background()
	if err := client.Verify(ctx); err != nil {
		e.log.Warn("credential verify failed, marking expired", "err", err)
		_ = e.orm.UpdateAccountStatus(acc.ID, orm.AccountStatusExpired)
		return nil, fmt.Errorf("credential verification failed: %w", err)
	}

	_ = e.orm.UpdateAccountStatus(acc.ID, orm.AccountStatusActive)
	return client, nil
}

// fetchData concurrently pulls buy history, sell history, and balance.
func (e *Engine) fetchData(client platform.Client, lastSyncAt *int64) (
	[]platform.TradeRecord, []platform.TradeRecord, *platform.Balance, []string,
) {
	since := int64(0)
	if lastSyncAt != nil {
		since = *lastSyncAt * 1000 // DB stores seconds, platform interface wants ms
	}
	e.log.Debug("fetching data", "since", utils.SecondsToDateTime(since, time.DateTime))

	var (
		buys, sells []platform.TradeRecord
		balance     *platform.Balance
		mu          sync.Mutex
		warnings    []string
		wg          sync.WaitGroup
	)
	ctx := context.Background()

	wg.Add(3)

	go func() {
		defer wg.Done()
		b, err := client.GetBuyHistory(ctx, platform.WithSince(since), platform.WithTradeState(platform.TradeStateCompleted))
		mu.Lock()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("buy history: %v", err))
		}
		e.log.Debug("buy history fetched", "count", len(b), "err", err)
		buys = b
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		s, err := client.GetSellHistory(ctx, platform.WithSince(since), platform.WithTradeState(platform.TradeStateCompleted))
		mu.Lock()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("sell history: %v", err))
		}
		e.log.Debug("sell history fetched", "count", len(s), "err", err)
		sells = s
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		b, err := client.GetBalance(ctx)
		mu.Lock()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("balance: %v", err))
		}
		e.log.Debug("balance fetched", "balance", b, "err", err)
		balance = b
		mu.Unlock()
	}()

	wg.Wait()
	return buys, sells, balance, warnings
}

// persistTrades converts platform records to models and saves them.
// Individual trade failures are logged as warnings and skipped.
func (e *Engine) persistTrades(accountID uint, source string, buys, sells []platform.TradeRecord) int {
	count := 0

	for _, b := range buys {
		t := toTradeModel(b, accountID, source, model.DirectionBuy)
		if err := e.orm.CreateTrade(&t); err != nil {
			e.log.Warn("create buy trade failed", "external_id", t.ExternalID, "err", err)
			continue
		}
		_ = e.orm.UpsertInventory(&model.InventoryItem{
			AccountID:  accountID,
			CS2Item:    t.CS2Item,
			BuyTradeID: t.ID,
			Quantity:   t.Quantity,
			Status:     "in_inventory",
		})
		count++
	}

	for _, s := range sells {
		t := toTradeModel(s, accountID, source, model.DirectionSell)
		if err := e.orm.CreateTrade(&t); err != nil {
			e.log.Warn("create sell trade failed", "external_id", t.ExternalID, "err", err)
			continue
		}
		count++
	}

	return count
}

// updateAccountMeta persists balance and sync timestamp.
func (e *Engine) updateAccountMeta(accountID uint, balance *platform.Balance, maxTradeMs int64, hasWarnings bool) {
	syncAt := time.Now().Unix()
	if hasWarnings {
		if maxTradeMs > 0 {
			syncAt = maxTradeMs / 1000
		} else {
			return
		}
	}
	if balance != nil {
		_ = e.orm.UpdateAccountBalanceAndSyncTime(accountID, int64(balance.Available*100), int64(balance.Purchase*100), syncAt)
	} else {
		_ = e.orm.UpdateAccountBalanceAndSyncTime(accountID, 0, 0, syncAt)
	}
}

func toTradeModel(r platform.TradeRecord, accountID uint, source, tradeType string) model.TradeRecord {
	return model.TradeRecord{
		AccountID:    accountID,
		CS2Item:      r.CS2Item,
		TradeType:    tradeType,
		Quantity:     r.Quantity,
		UnitPrice:    r.UnitPrice,
		TotalPrice:   r.TotalPrice,
		Fee:          r.Fee,
		TradeAt:      r.TradeAt,
		Source:       source,
		State:        r.State,
		StateText:    r.StateText,
		TransactTime: &r.TransactTime,
		TradeOfferID: r.TradeOfferID,
		ExternalID:   r.ExternalID,
	}
}

// aggregateBulkTrades merges fungible bulk-item trades (cases, stickers, keys,
// capsules, etc.) into per-day per-item records. Items without exterior wear or
// paint attributes are treated as bulk.
func aggregateBulkTrades(trades []platform.TradeRecord, platformName, direction string) []platform.TradeRecord {
	var regular, bulk []platform.TradeRecord
	for _, t := range trades {
		if isBulkItem(t.CS2Item) {
			bulk = append(bulk, t)
		} else {
			regular = append(regular, t)
		}
	}
	if len(bulk) <= 1 {
		return trades
	}

	type aggKey struct {
		GoodsID int
		Date    string // "2006-01-02" in UTC
	}
	groups := make(map[aggKey][]platform.TradeRecord)
	for _, t := range bulk {
		day := time.UnixMilli(t.TradeAt).UTC().Format("2006-01-02")
		key := aggKey{GoodsID: t.GoodsID, Date: day}
		groups[key] = append(groups[key], t)
	}

	var aggregated []platform.TradeRecord
	for key, records := range groups {
		if len(records) == 1 {
			regular = append(regular, records[0])
			continue
		}
		var qty, totalPrice, fee int64
		for _, r := range records {
			qty += r.Quantity
			totalPrice += r.TotalPrice
			fee += r.Fee
		}
		first := records[0]
		first.Quantity = qty
		first.TotalPrice = totalPrice
		first.UnitPrice = totalPrice / qty
		first.Fee = fee
		first.ExternalID = fmt.Sprintf("%s-%s-agg-%d-%s", platformName, direction, key.GoodsID, key.Date)
		aggregated = append(aggregated, first)
	}

	return append(regular, aggregated...)
}

func maxTradeAt(buys, sells []platform.TradeRecord) int64 {
	var max int64
	for _, t := range buys {
		if t.TradeAt > max {
			max = t.TradeAt
		}
	}
	for _, t := range sells {
		if t.TradeAt > max {
			max = t.TradeAt
		}
	}
	return max
}

var Module = fx.Module("sync",
	logfx.WithComponent("sync"),
	fx.Provide(NewEngine),
)
