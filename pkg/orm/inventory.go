package orm

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

// DailyBuyRow is a denormalized row for daily-buy queries, assembled from an inventory item and its buy trade.
type DailyBuyRow struct {
	ItemName       string
	Exterior       string
	Quantity       int64
	BuyPrice       int64
	BuyAt          int64
	Source         string
	Status         string
	MarketHashName string
	CsqaqGoodsID   int
}

func (o *ormImpl) UpsertInventory(item *model.InventoryItem) error {
	return o.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "asset_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"item_name", "quantity", "status", "listed_price", "listed_at", "exterior", "paint_seed", "updated_at"}),
	}).Create(item).Error
}

func (o *ormImpl) RemoveInventoryByAssetID(accountID uint, assetID string) error {
	return o.db.Unscoped().Where("account_id = ? AND asset_id = ?", accountID, assetID).
		Delete(&model.InventoryItem{}).Error
}

func (o *ormImpl) FindInventoryByAccount(accountID uint, status string) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	q := o.db.Where("account_id = ?", accountID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Preload("BuyTrade").Order("updated_at DESC").Find(&items).Error
	return items, err
}

func (o *ormImpl) FindInventoryByAssetID(accountID uint, assetID string) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := o.db.Where("account_id = ? AND asset_id = ?", accountID, assetID).Preload("BuyTrade").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

// FindInventoryGroupKeys returns paginated distinct (item_name, exterior) pairs for an account.
// Pass accountID=0 to query across all accounts.
func (o *ormImpl) FindInventoryGroupKeys(accountID uint, status, weaponType string, offset, limit int, sortBy, sortDir string) ([]InventoryGroupKey, int64, error) {
	var total int64
	q := o.db.Model(&model.InventoryItem{})
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	} else {
		q = q.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if weaponType != "" {
		q = q.Where("weapon_type = ?", weaponType)
	}
	if err := q.Select("COUNT(DISTINCT item_name || '|' || COALESCE(exterior, ''))").Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	var keys []InventoryGroupKey
	q2 := o.db.Model(&model.InventoryItem{}).Select("item_name, COALESCE(exterior, '') AS exterior")
	if accountID != 0 {
		q2 = q2.Where("account_id = ?", accountID)
	} else {
		q2 = q2.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	if status != "" {
		q2 = q2.Where("status = ?", status)
	}
	if weaponType != "" {
		q2 = q2.Where("weapon_type = ?", weaponType)
	}
	err := q2.Group("item_name, exterior").Order(sortBy + " " + sortDir).
		Offset(offset).Limit(limit).
		Find(&keys).Error
	return keys, total, err
}

// FindInventoryByGroupKeys returns inventory items matching the given (item_name, exterior) pairs.
// Pass accountID=0 to query across all accounts.
func (o *ormImpl) FindDailyBuys(accountID uint) ([]DailyBuyRow, error) {
	q := o.db.Model(&model.InventoryItem{}).
		Where("status IN ?", []string{model.InventoryStatusInInventory, model.InventoryStatusListed}).
		Preload("BuyTrade")
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	}

	var items []model.InventoryItem
	if err := q.Order("updated_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	rows := make([]DailyBuyRow, 0, len(items))
	for _, it := range items {
		if it.BuyTrade == nil {
			continue
		}
		rows = append(rows, DailyBuyRow{
			ItemName:       it.ItemName,
			Exterior:       it.Exterior,
			Quantity:       it.Quantity,
			BuyPrice:       it.BuyTrade.UnitPrice,
			BuyAt:          it.BuyTrade.TradeAt,
			Source:         it.BuyTrade.Source,
			Status:         it.Status,
			MarketHashName: it.MarketHashName,
			CsqaqGoodsID:   it.CsqaqGoodsID,
		})
	}
	return rows, nil
}

func (o *ormImpl) FindInventoryByGroupKeys(accountID uint, keys []InventoryGroupKey) ([]model.InventoryItem, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	var conditions []string
	var args []any
	for _, k := range keys {
		conditions = append(conditions, "(item_name = ? AND COALESCE(exterior, '') = ?)")
		args = append(args, k.ItemName, k.Exterior)
	}

	var items []model.InventoryItem
	q := o.db.Where("("+strings.Join(conditions, " OR ")+")", args...)
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	} else {
		q = q.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	err := q.Preload("BuyTrade").Order("updated_at DESC").Find(&items).Error
	return items, err
}
