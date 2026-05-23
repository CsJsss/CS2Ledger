package orm

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

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

// FindInventoryGroupNames returns paginated distinct item names for an account.
// Pass accountID=0 to query across all accounts.
func (o *ormImpl) FindInventoryGroupNames(accountID uint, status, weaponType string, offset, limit int, sortBy, sortDir string) ([]string, int64, error) {
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
	if err := q.Select("COUNT(DISTINCT item_name)").Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	var names []string
	q2 := o.db.Model(&model.InventoryItem{}).Select("item_name")
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
	err := q2.Group("item_name").Order(sortBy+" "+sortDir).
		Offset(offset).Limit(limit).
		Pluck("item_name", &names).Error
	return names, total, err
}

// FindInventoryByItemNames returns inventory items matching the given item names.
// Pass accountID=0 to query across all accounts.
func (o *ormImpl) FindInventoryByItemNames(accountID uint, itemNames []string) ([]model.InventoryItem, error) {
	if len(itemNames) == 0 {
		return nil, nil
	}
	var items []model.InventoryItem
	q := o.db.Where("item_name IN ?", itemNames)
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	} else {
		q = q.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	err := q.Preload("BuyTrade").Order("updated_at DESC").Find(&items).Error
	return items, err
}
