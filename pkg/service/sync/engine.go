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
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type Engine struct {
	log     *logfx.Logger
	factory *factory.Factory
	orm     orm.ORMInterface
	pnlSvc  pnl.PnlInterface
}

func NewEngine(
	log *logfx.Logger,
	f *factory.Factory,
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
	acc, err := e.orm.FindAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	e.log.Info("sync started", "account_id", accountID, "platform", acc.Platform, "name", acc.Name)

	client, err := e.factory.New(acc.Platform, acc.Cookie)
	if err != nil {
		e.log.Error("sync: create client failed", "err", err)
		return nil, fmt.Errorf("create client: %w", err)
	}

	syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := client.Verify(syncCtx); err != nil {
		e.log.Warn("sync: credential verify failed, marking expired", "err", err)
		_ = e.orm.UpdateAccountStatus(accountID, "expired")
		return nil, fmt.Errorf("credential verification failed: %w", err)
	}

	_ = e.orm.UpdateAccountStatus(accountID, "active")

	since := int64(0)
	if acc.LastSyncAt != nil {
		since = *acc.LastSyncAt * 1000 // lastSyncAt is seconds, interface wants ms
	}
	e.log.Debug("sync: fetching data", "since", since)

	result := &SyncResult{}
	var fetchErrors []string

	var buys, sells []platform.TradeRecord
	var balance *platform.Balance
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		buys, err = client.FetchBuyHistory(syncCtx, since)
		mu.Lock()
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Sprintf("buy history: %v", err))
		}
		e.log.Debug("sync: buy history fetched", "count", len(buys), "err", err)
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		sells, err = client.FetchSellHistory(syncCtx, since)
		mu.Lock()
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Sprintf("sell history: %v", err))
		}
		e.log.Debug("sync: sell history fetched", "count", len(sells), "err", err)
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		balance, err = client.FetchBalance(syncCtx)
		mu.Lock()
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Sprintf("balance: %v", err))
		}
		e.log.Debug("sync: balance fetched", "balance", balance, "err", err)
		mu.Unlock()
	}()

	wg.Wait()
	result.Warnings = fetchErrors

	e.log.Debug("sync: saving trades", "buys", len(buys), "sells", len(sells))

	for _, b := range buys {
		t := toTradeModel(b, accountID, "buy")
		if err := e.orm.CreateTrade(&t); err != nil {
			e.log.Warn("sync: create buy trade failed", "external_id", t.ExternalID, "err", err)
			continue
		}
		_ = e.orm.UpsertInventory(&model.InventoryItem{
			AccountID:  accountID,
			AssetID:    t.AssetID,
			ItemName:   t.ItemName,
			BuyTradeID: t.ID,
			Status:     "in_inventory",
		})
		result.NewTrades++
	}

	for _, s := range sells {
		t := toTradeModel(s, accountID, "sell")
		if err := e.orm.CreateTrade(&t); err != nil {
			e.log.Warn("sync: create sell trade failed", "external_id", t.ExternalID, "err", err)
			continue
		}
		result.NewTrades++
	}

	pnlCount, pnlErr := e.pnlSvc.ProcessPending(accountID)
	result.NewPnl = pnlCount
	e.log.Debug("sync: PNL processed", "matched", pnlCount, "err", pnlErr)

	now := time.Now().Unix()
	if balance != nil {
		_ = e.orm.UpdateAccountBalanceAndSyncTime(accountID, balance.Available, balance.Purchase, now)
	} else {
		_ = e.orm.UpdateAccountBalanceAndSyncTime(accountID, 0, 0, now)
	}

	e.log.Info("sync: completed",
		"new_trades", result.NewTrades,
		"new_pnl", result.NewPnl,
		"warnings", len(fetchErrors),
	)

	return result, nil
}

func toTradeModel(r platform.TradeRecord, accountID uint, tradeType string) model.TradeRecord {
	return model.TradeRecord{
		AccountID:  accountID,
		AssetID:    r.AssetID,
		ItemName:   r.ItemName,
		TradeType:  tradeType,
		Quantity:   r.Quantity,
		UnitPrice:  r.UnitPrice,
		TotalPrice: r.TotalPrice,
		Fee:        r.Fee,
		TradeAt:    r.TradeAt,
		ExternalID: r.ExternalID,
	}
}

var Module = fx.Module("sync",
	logfx.WithComponent("sync"),
	fx.Provide(NewEngine),
)
