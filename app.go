package main

import (
	"context"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/service"
	"github.com/CsJsss/CS2Ledger/pkg/service/inventory"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/service/sync"
	"github.com/CsJsss/CS2Ledger/pkg/service/trade"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type DashboardSummary struct {
	TotalNetWorth     int64 `json:"totalNetWorth"`
	InventoryCount    int64 `json:"inventoryCount"`
	CompletedTrades   int64 `json:"completedTrades"`
	TotalRentalIncome int64 `json:"totalRentalIncome"`
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

func (a *App) CreateAccount(name, platform, cookie string) (*model.Account, error) {
	return a.svc.Account().Create(name, platform, cookie)
}

func (a *App) UpdateAccount(acc *model.Account) error {
	return a.svc.Account().Update(acc)
}

func (a *App) UpdateAccountInfo(id uint, name string, cookie string) error {
	return a.svc.Account().UpdateInfo(id, name, cookie)
}

func (a *App) DeleteAccount(id uint) error {
	return a.svc.Account().Delete(id)
}

func (a *App) SyncAccount(accountID uint) (*sync.SyncResult, error) {
	return a.syncEngine.SyncAccount(accountID)
}

func (a *App) GetInventory(accountID uint, status string) ([]model.InventoryItem, error) {
	return a.svc.Inventory().List(accountID, status)
}

func (a *App) GetItemDetail(accountID uint, assetID string) (*inventory.ItemDetail, error) {
	return a.svc.Inventory().GetItemDetail(accountID, assetID)
}

func (a *App) GetCompletedTrades(accountID uint) ([]trade.CompletedTradeView, error) {
	return a.svc.Trade().ListCompletedTrades(accountID)
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
	ds := &DashboardSummary{}
	for _, acc := range accounts {
		inv, _ := a.svc.Inventory().List(acc.ID, "")
		ds.InventoryCount += int64(len(inv))
		summary, _ := a.svc.Pnl().GetSummary(acc.ID)
		if summary != nil {
			ds.CompletedTrades += summary.TotalTrades
			ds.TotalNetWorth += summary.TotalNetPl
		}
	}
	return ds, nil
}

func (a *App) GetRentalHistory(accountID uint, assetID string) ([]model.RentalRecord, error) {
	return a.svc.Rental().ListByAsset(accountID, assetID)
}
