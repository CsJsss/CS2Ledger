package model

import "gorm.io/gorm"

type TradeRecord struct {
	gorm.Model
	AccountID         uint   `gorm:"not null;index:idx_trades_account" json:"accountId"`
	AssetID           string `gorm:"not null;index:idx_trades_asset" json:"assetId"`
	ItemName          string `gorm:"not null" json:"itemName"`
	TradeType         string `gorm:"not null;index:idx_trades_type" json:"tradeType"`
	Quantity          int64  `gorm:"not null;default:1" json:"quantity"`
	UnitPrice         int64  `gorm:"not null" json:"unitPrice"`
	TotalPrice        int64  `gorm:"not null" json:"totalPrice"`
	Fee               int64  `gorm:"not null;default:0" json:"fee"`
	TradeAt           int64  `gorm:"not null" json:"tradeAt"`
	Source            string `gorm:"not null;default:platform" json:"source"`
	ExternalID        string `gorm:"uniqueIndex:idx_trades_external" json:"externalId"`
	MatchedBuyTradeID *uint  `gorm:"index:idx_trades_matched" json:"matchedBuyTradeId"`
	Remark            string `json:"remark"`

	Account    *Account     `gorm:"foreignKey:AccountID" json:"-"`
	MatchedBuy *TradeRecord `gorm:"foreignKey:MatchedBuyTradeID" json:"-"`
}

func (TradeRecord) TableName() string { return "trade_records" }
