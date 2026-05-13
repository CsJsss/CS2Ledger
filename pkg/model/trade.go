package model

import "gorm.io/gorm"

const (
	DirectionBuy  = "buy"
	DirectionSell = "sell"
)

type TradeRecord struct {
	gorm.Model
	CS2Item           `gorm:"embedded"`
	AccountID         uint   `gorm:"not null;index:idx_trades_account" json:"accountId"`
	TradeType         string `gorm:"not null;index:idx_trades_type" json:"tradeType"`
	Quantity          int64  `gorm:"not null;default:1" json:"quantity"`
	UnitPrice         int64  `gorm:"not null" json:"unitPrice"`
	TotalPrice        int64  `gorm:"not null" json:"totalPrice"`
	Fee               int64  `gorm:"not null;default:0" json:"fee"`
	TradeAt           int64  `gorm:"not null" json:"tradeAt"`
	Source            string `gorm:"not null" json:"source"`
	State             string `gorm:"not null;default:SUCCESS" json:"state"`
	StateText         string `json:"stateText"`
	TransactTime      *int64 `json:"transactTime"`
	TradeOfferID      string `json:"tradeOfferId"`
	ExternalID        string `gorm:"uniqueIndex:idx_trades_external" json:"externalId"`
	MatchedBuyTradeID *uint  `gorm:"index:idx_trades_matched" json:"matchedBuyTradeId"`
	ConsumedQuantity  int64  `gorm:"not null;default:0" json:"consumedQuantity"`
	Remark            string `json:"remark"`

	Account    *Account     `gorm:"foreignKey:AccountID" json:"-"`
	MatchedBuy *TradeRecord `gorm:"foreignKey:MatchedBuyTradeID" json:"-"`
}

func (TradeRecord) TableName() string { return "trade_records" }
