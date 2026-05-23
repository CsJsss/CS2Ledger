package orm

import "github.com/CsJsss/CS2Ledger/pkg/model"

type AccountInterface interface {
	CreateAccount(*model.Account) error
	ListAccounts() ([]model.Account, error)
	FindAccountByID(id uint) (*model.Account, error)
	UpdateAccount(*model.Account) error
	DeleteAccount(id uint) error
	UpdateAccountInfo(id uint, name string, cookie string) error
	UpdateAccountStatus(id uint, status string) error
	UpdateAccountBalanceAndSyncTime(id uint, available, frozen, instant, purchase int64, syncAt int64) error
}

type TradeInterface interface {
	CreateTrade(*model.TradeRecord) error
	FindTradesByAccount(accountID uint, tradeType string, limit int) ([]model.TradeRecord, error)
	FindSellsWithMatchedBuy(accountID uint) ([]model.TradeRecord, error)
	FindUnmatchedSells(accountID uint) ([]model.TradeRecord, error)
	FindUnmatchedBuysByItem(itemName string, paintSeed, paintIndex int, paintWear float64) ([]model.TradeRecord, error)
	FindMatchedBuyForSell(sellID uint) (*model.TradeRecord, error)
	SetMatchedBuy(sellTradeID, buyTradeID uint) error
	IncrementConsumedQty(buyID uint, qty int64) error
	FindAllSells(tradeType string) ([]model.TradeRecord, error)
	FindAllBuys(tradeType string) ([]model.TradeRecord, error)
	FindEarliestUnmatchedBuy(itemName, exterior string, paintSeed, paintIndex int, paintWear float64, beforeTime int64) (*model.TradeRecord, error)
	ClearAllMatches() error
	RebuildInventory() error
	FindCompletedTradeGroupNames(accountID uint, offset, limit int, sortBy, sortDir string) ([]string, int64, error)
	FindSellsByItemNames(accountID uint, itemNames []string) ([]model.TradeRecord, error)
	FindTradeRecordsByIDs(ids []uint) ([]model.TradeRecord, error)
}

type InventoryGroupKey struct {
	ItemName string
	Exterior string
}

type InventoryInterface interface {
	UpsertInventory(*model.InventoryItem) error
	RemoveInventoryByAssetID(accountID uint, assetID string) error
	FindInventoryByAccount(accountID uint, status string) ([]model.InventoryItem, error)
	FindInventoryByAssetID(accountID uint, assetID string) (*model.InventoryItem, error)
	FindInventoryGroupKeys(accountID uint, status, weaponType string, offset, limit int, sortBy, sortDir string) ([]InventoryGroupKey, int64, error)
	FindInventoryByGroupKeys(accountID uint, keys []InventoryGroupKey) ([]model.InventoryItem, error)
}

type PnlInterface interface {
	UpsertDailyPnl(accountID uint, tradeAt int64, grossPl, fee, netPl int64) error
	FindPnlByAccount(accountID uint) ([]model.PnlDaily, error)
	ClearAllPnl() error
	ReplaceAllPnl(records []model.PnlDaily) error
}

type RentalInterface interface {
	CreateRental(*model.RentalRecord) error
	FindRentalsByAssetID(accountID uint, assetID string) ([]model.RentalRecord, error)
	FindRentalsByAccount(accountID uint) ([]model.RentalRecord, error)
}

type ORMInterface interface {
	AccountInterface
	TradeInterface
	InventoryInterface
	PnlInterface
	RentalInterface
}
