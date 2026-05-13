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
