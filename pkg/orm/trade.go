package orm

import (
	"strings"

	"gorm.io/gorm"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

func (o *ormImpl) CreateTrade(t *model.TradeRecord) error {
	t.ItemName = strings.TrimSpace(t.ItemName)
	if t.ExternalID == "" {
		return o.db.Create(t).Error
	}
	return o.db.Where("account_id = ? AND external_id = ?", t.AccountID, t.ExternalID).
		Assign(t).FirstOrCreate(t).Error
}

func (o *ormImpl) FindTradesByAccount(accountID uint, tradeType string, limit int) ([]model.TradeRecord, error) {
	var trades []model.TradeRecord
	q := o.db.Where("account_id = ?", accountID)
	if tradeType != "" {
		q = q.Where("trade_type = ?", tradeType)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Order("trade_at DESC").Find(&trades).Error
	return trades, err
}

func (o *ormImpl) FindSellsWithMatchedBuy(accountID uint) ([]model.TradeRecord, error) {
	var sells []model.TradeRecord
	err := o.db.Where(
		"account_id = ? AND trade_type = 'sell' AND matched_buy_trade_id IS NOT NULL",
		accountID,
	).Order("trade_at DESC").Find(&sells).Error
	return sells, err
}

func (o *ormImpl) FindUnmatchedSells(accountID uint) ([]model.TradeRecord, error) {
	var sells []model.TradeRecord
	err := o.db.Where(
		"account_id = ? AND trade_type = 'sell' AND matched_buy_trade_id IS NULL",
		accountID,
	).Order("trade_at ASC").Find(&sells).Error
	return sells, err
}

func (o *ormImpl) FindUnmatchedBuysByItem(itemName string, paintSeed, paintIndex int, paintWear float64) ([]model.TradeRecord, error) {
	var buys []model.TradeRecord
	err := o.db.Where(
		"trade_type = 'buy' AND item_name = ? AND paint_seed = ? AND paint_index = ? AND paint_wear BETWEEN ? AND ? AND consumed_quantity < quantity",
		itemName, paintSeed, paintIndex, paintWear-0.0001, paintWear+0.0001,
	).Order("trade_at ASC").Find(&buys).Error
	return buys, err
}

func (o *ormImpl) FindMatchedBuyForSell(sellID uint) (*model.TradeRecord, error) {
	var sell model.TradeRecord
	if err := o.db.Where("id = ?", sellID).First(&sell).Error; err != nil {
		return nil, err
	}
	if sell.MatchedBuyTradeID == nil {
		return nil, nil
	}
	var buy model.TradeRecord
	if err := o.db.Where("id = ?", *sell.MatchedBuyTradeID).First(&buy).Error; err != nil {
		return nil, err
	}
	return &buy, nil
}

func (o *ormImpl) SetMatchedBuy(sellTradeID, buyTradeID uint) error {
	return o.db.Model(&model.TradeRecord{}).Where("id = ?", sellTradeID).
		Update("matched_buy_trade_id", buyTradeID).Error
}

func (o *ormImpl) IncrementConsumedQty(buyID uint, qty int64) error {
	return o.db.Model(&model.TradeRecord{}).Where("id = ?", buyID).
		Update("consumed_quantity", gorm.Expr("consumed_quantity + ?", qty)).Error
}

func (o *ormImpl) FindAllSells(tradeType string) ([]model.TradeRecord, error) {
	var trades []model.TradeRecord
	err := o.db.Where("trade_type = ?", tradeType).Order("trade_at ASC").Find(&trades).Error
	return trades, err
}

func (o *ormImpl) FindAllBuys(tradeType string) ([]model.TradeRecord, error) {
	var trades []model.TradeRecord
	err := o.db.Where("trade_type = ?", tradeType).Order("trade_at ASC").Find(&trades).Error
	return trades, err
}

func (o *ormImpl) FindEarliestUnmatchedBuy(itemName, exterior string, paintSeed, paintIndex int, paintWear float64, beforeTime int64) (*model.TradeRecord, error) {
	var buys []model.TradeRecord
	err := o.db.Where(
		"trade_type = 'buy' AND item_name = ? AND (exterior = ? OR exterior = '' OR ? = '') AND paint_seed = ? AND paint_index = ? AND paint_wear BETWEEN ? AND ? AND trade_at <= ? AND consumed_quantity < quantity",
		itemName, exterior, exterior, paintSeed, paintIndex, paintWear-0.0001, paintWear+0.0001, beforeTime,
	).Order("trade_at ASC").Limit(1).Find(&buys).Error
	if err != nil || len(buys) == 0 {
		return nil, nil
	}
	return &buys[0], nil
}

func (o *ormImpl) ClearAllMatches() error {
	if err := o.db.Model(&model.TradeRecord{}).Where("trade_type = 'buy'").
		Update("consumed_quantity", 0).Error; err != nil {
		return err
	}
	return o.db.Model(&model.TradeRecord{}).
		Where("trade_type = 'sell'").
		Update("matched_buy_trade_id", nil).Error
}

func (o *ormImpl) RebuildInventory() error {
	return o.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1=1").Delete(&model.InventoryItem{}).Error; err != nil {
			return err
		}
		return tx.Exec(`
			INSERT OR IGNORE INTO inventory
				(account_id, asset_id, buy_trade_id, quantity, status,
				 item_name, market_hash_name, icon_url, class_id, instance_id,
				 goods_id, category_group, paint_wear, paint_seed, paint_index,
				 exterior, rarity, weapon_type, weapon_name, quality, series,
				 itemset, weapon_case, custom, stickers, keychains, tradable_unfrozen_time,
				 created_at, updated_at)
			SELECT
				account_id, asset_id, id,
				quantity - consumed_quantity,
				'in_inventory',
				item_name, market_hash_name, icon_url, class_id, instance_id,
				goods_id, category_group, paint_wear, paint_seed, paint_index,
				exterior, rarity, weapon_type, weapon_name, quality, series,
				itemset, weapon_case, custom, stickers, keychains, tradable_unfrozen_time,
				DATETIME('now'), DATETIME('now')
			FROM trade_records
			WHERE trade_type = 'buy'
				AND quantity - consumed_quantity > 0
		`).Error
	})
}

// FindCompletedTradeGroupNames returns paginated distinct item names for completed trades.
// Pass accountID=0 to query across all accounts.
func (o *ormImpl) FindCompletedTradeGroupNames(accountID uint, offset, limit int, sortBy, sortDir string) ([]string, int64, error) {
	var total int64
	q := o.db.Model(&model.TradeRecord{}).
		Where("trade_type = 'sell' AND matched_buy_trade_id IS NOT NULL")
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	} else {
		q = q.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	if err := q.Select("COUNT(DISTINCT item_name)").Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	var names []string
	q2 := o.db.Model(&model.TradeRecord{}).
		Select("item_name").
		Where("trade_type = 'sell' AND matched_buy_trade_id IS NOT NULL")
	if accountID != 0 {
		q2 = q2.Where("account_id = ?", accountID)
	} else {
		q2 = q2.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	err := q2.Group("item_name").Order(sortBy+" "+sortDir).
		Offset(offset).Limit(limit).
		Pluck("item_name", &names).Error
	return names, total, err
}

// FindSellsByItemNames returns matched sells for the given item names.
// Pass accountID=0 to query across all accounts.
func (o *ormImpl) FindSellsByItemNames(accountID uint, itemNames []string) ([]model.TradeRecord, error) {
	if len(itemNames) == 0 {
		return nil, nil
	}
	var sells []model.TradeRecord
	q := o.db.Where(
		"item_name IN ? AND trade_type = 'sell' AND matched_buy_trade_id IS NOT NULL",
		itemNames,
	)
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	} else {
		q = q.Where("account_id IN (SELECT id FROM accounts WHERE deleted_at IS NULL)")
	}
	err := q.Order("trade_at DESC").Find(&sells).Error
	return sells, err
}

func (o *ormImpl) FindTradeRecordsByIDs(ids []uint) ([]model.TradeRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var records []model.TradeRecord
	err := o.db.Where("id IN ?", ids).Find(&records).Error
	return records, err
}
