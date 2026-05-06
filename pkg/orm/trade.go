package orm

import "github.com/CsJsss/CS2Ledger/pkg/model"

func (o *ormImpl) CreateTrade(t *model.TradeRecord) error {
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
	err := o.db.Where("account_id = ? AND trade_type = 'sell' AND matched_buy_trade_id IS NOT NULL",
		accountID).Order("trade_at DESC").Find(&sells).Error
	return sells, err
}

func (o *ormImpl) FindUnmatchedSells(accountID uint) ([]model.TradeRecord, error) {
	var sells []model.TradeRecord
	err := o.db.Where("account_id = ? AND trade_type = 'sell' AND matched_buy_trade_id IS NULL",
		accountID).Order("trade_at ASC").Find(&sells).Error
	return sells, err
}

func (o *ormImpl) FindUnmatchedBuys(accountID uint, assetID string) ([]model.TradeRecord, error) {
	var buys []model.TradeRecord
	err := o.db.Where("account_id = ? AND asset_id = ? AND trade_type = 'buy'",
		accountID, assetID).Order("trade_at ASC").Find(&buys).Error
	return buys, err
}

func (o *ormImpl) SetMatchedBuy(sellTradeID, buyTradeID uint) error {
	return o.db.Model(&model.TradeRecord{}).Where("id = ?", sellTradeID).
		Update("matched_buy_trade_id", buyTradeID).Error
}

func (o *ormImpl) CountMatchedSellsForBuy(buyTradeID uint) (int64, error) {
	var count int64
	err := o.db.Model(&model.TradeRecord{}).
		Where("matched_buy_trade_id = ?", buyTradeID).Count(&count).Error
	return count, err
}
