package main

import (
	"context"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/service"
	"github.com/CsJsss/CS2Ledger/pkg/service/inventory"
	"github.com/CsJsss/CS2Ledger/pkg/service/market"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/service/sync"
	"github.com/CsJsss/CS2Ledger/pkg/service/trade"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type DashboardSummary struct {
	RealizedPl            int64  `json:"realizedPl"`
	InventoryCount        int64  `json:"inventoryCount"`
	InventoryCost         int64  `json:"inventoryCost"`
	InventoryMarketValue  int64  `json:"inventoryMarketValue"`
	PriceSource           string `json:"priceSource"`
	CompletedTrades       int64  `json:"completedTrades"`
	TotalRentalIncome     int64  `json:"totalRentalIncome"`
	TotalAvailableBalance int64  `json:"totalAvailableBalance"`
	TotalFrozenBalance    int64  `json:"totalFrozenBalance"`
	TotalInstantBalance   int64  `json:"totalInstantBalance"`
	TotalPurchaseBalance  int64  `json:"totalPurchaseBalance"`
}

type App struct {
	ctx        context.Context
	log        *logfx.Logger
	svc        *service.Service
	syncEngine *sync.Engine
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.log.Info("Wails started")
	a.svc.Market().StartAutoRefresh(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.log.Info("Wails shutting down")
}

func NewApp(
	log *logfx.Logger,
	svc *service.Service,
	syncEngine *sync.Engine,
) *App {
	return &App{
		log:        log,
		svc:        svc,
		syncEngine: syncEngine,
	}
}

func (a *App) GetAccounts() ([]model.Account, error) {
	return a.svc.Account().List()
}

func (a *App) CreateAccount(name, platformName, cookie string) (*model.Account, error) {
	acc, err := a.svc.Account().Create(name, platformName, cookie)
	if err == nil && platformName == platform.PlatformCsqaq {
		// New csqaq account added — trigger immediate price refresh
		a.svc.Market().StartAutoRefresh(a.ctx)
	}
	return acc, err
}

func (a *App) UpdateAccount(acc *model.Account) error {
	return a.svc.Account().Update(acc)
}

func (a *App) UpdateAccountInfo(id uint, name string, cookie string) error {
	if err := a.svc.Account().UpdateInfo(id, name, cookie); err != nil {
		return err
	}
	// If csqaq account's token may have changed, trigger refresh
	accs, _ := a.svc.Account().List()
	for _, acc := range accs {
		if acc.ID == id && acc.Platform == platform.PlatformCsqaq {
			a.svc.Market().StartAutoRefresh(a.ctx)
			break
		}
	}
	return nil
}

func (a *App) DeleteAccount(id uint) error {
	return a.svc.Account().Delete(id)
}

func (a *App) SyncAccount(accountID uint) (*sync.SyncResult, error) {
	return a.syncEngine.SyncAccount(accountID)
}

func (a *App) GetInventory(accountID uint, status, weaponType string, page, pageSize int, sortBy, sortDir string) (*inventory.PaginatedGroups, error) {
	return a.svc.Inventory().ListGroups(accountID, status, weaponType, page, pageSize, sortBy, sortDir)
}

func (a *App) GetItemDetail(accountID uint, assetID string) (*inventory.ItemDetail, error) {
	return a.svc.Inventory().GetItemDetail(accountID, assetID)
}

func (a *App) GetCompletedTrades(accountID uint, page, pageSize int, sortBy, sortDir string) (*trade.PaginatedGroups, error) {
	return a.svc.Trade().ListCompletedTradeGroups(accountID, page, pageSize, sortBy, sortDir)
}

func (a *App) GetCompletedTradesSummary(accountID uint) (*trade.CompletedTradesSummary, error) {
	s, err := a.svc.Pnl().GetSummary(accountID)
	if err != nil {
		return nil, err
	}
	return &trade.CompletedTradesSummary{
		TotalTrades:  s.TotalTrades,
		TotalGrossPl: s.TotalGrossPl,
		TotalFee:     s.TotalFee,
		TotalNetPl:   s.TotalNetPl,
	}, nil
}

func (a *App) GetUnmatchedSells(accountID uint) ([]model.TradeRecord, error) {
	return a.svc.Trade().ListUnmatchedSells(accountID)
}

func (a *App) GetPnlSummary(accountID uint) (*pnl.PnlSummaryView, error) {
	return a.svc.Pnl().GetSummary(accountID)
}

func (a *App) GetMonthlyBreakdown(accountID uint, year int) ([]pnl.MonthlyPLView, error) {
	return a.svc.Pnl().GetMonthlyBreakdown(accountID, year)
}

func (a *App) GetDashboardSummary() (*DashboardSummary, error) {
	accounts, err := a.svc.Account().List()
	if err != nil {
		return nil, err
	}

	// Fetch market prices once for all accounts
	prices, _ := a.svc.Market().GetAllPrices()
	cfg := a.svc.Config()
	priceMap := make(map[string]int64, len(prices))
	for _, p := range prices {
		switch cfg.PriceSource {
		case "youpin":
			priceMap[p.MarketHashName] = int64(p.YoupinPrice * 100)
		case "steam":
			priceMap[p.MarketHashName] = int64(p.SteamPrice * 100)
		default:
			priceMap[p.MarketHashName] = int64(p.BuffPrice * 100)
		}
	}

	ds := &DashboardSummary{PriceSource: cfg.PriceSource}
	for _, acc := range accounts {
		inv, _ := a.svc.Inventory().List(acc.ID, "")
		ds.InventoryCount += int64(len(inv))

		for _, item := range inv {
			qty := item.Quantity
			if qty == 0 {
				qty = 1
			}
			if item.BuyTrade != nil {
				ds.InventoryCost += item.BuyTrade.UnitPrice * qty
			}
			if item.MarketHashName != "" {
				if mp, ok := priceMap[item.MarketHashName]; ok {
					ds.InventoryMarketValue += mp * qty
				}
			}
		}

		ds.TotalAvailableBalance += acc.AvailableBalance
		ds.TotalFrozenBalance += acc.FrozenBalance
		ds.TotalInstantBalance += acc.InstantBalance
		ds.TotalPurchaseBalance += acc.PurchaseBalance
		summary, _ := a.svc.Pnl().GetSummary(acc.ID)
		if summary != nil {
			ds.CompletedTrades += summary.TotalTrades
			ds.RealizedPl += summary.TotalNetPl
		}
	}
	return ds, nil
}

func (a *App) GetRentalHistory(accountID uint, assetID string) ([]model.RentalRecord, error) {
	return a.svc.Rental().ListByAsset(accountID, assetID)
}

type UserSettings struct {
	PriceSource   string `json:"priceSource"`
	PriceCacheTTL int    `json:"priceCacheTtl"`
}

func (a *App) GetMarketPrices() ([]platform.PriceInfo, error) {
	return a.svc.Market().GetAllPrices()
}

func (a *App) GetSettings() *UserSettings {
	cfg := a.svc.Config()
	return &UserSettings{
		PriceSource:   cfg.PriceSource,
		PriceCacheTTL: cfg.PriceCacheTTL,
	}
}

func (a *App) UpdateSettings(s *UserSettings) error {
	if s.PriceCacheTTL < 5 {
		s.PriceCacheTTL = 5
	}
	if s.PriceCacheTTL > 1440 {
		s.PriceCacheTTL = 1440
	}
	validSources := map[string]bool{"buff": true, "youpin": true, "steam": true}
	if !validSources[s.PriceSource] {
		s.PriceSource = "buff"
	}
	cfg := a.svc.Config()
	if err := cfg.UpdatePriceSettings(s.PriceSource, s.PriceCacheTTL); err != nil {
		return err
	}
	a.svc.Market().SetConfig(market.PriceConfig{
		PriceSource: s.PriceSource,
		CacheTTLMin: s.PriceCacheTTL,
	})
	a.svc.Inventory().SetPriceSource(s.PriceSource)
	return nil
}
