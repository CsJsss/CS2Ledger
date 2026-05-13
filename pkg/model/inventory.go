package model

import "gorm.io/gorm"

type InventoryItem struct {
	gorm.Model
	CS2Item     `gorm:"embedded"`
	AccountID   uint   `gorm:"not null" json:"accountId"`
	BuyTradeID  uint   `gorm:"not null" json:"buyTradeId"`
	Quantity    int64  `gorm:"not null;default:1" json:"quantity"`
	Status      string `gorm:"not null;default:in_inventory;index:idx_inventory_status" json:"status"`
	ListedPrice *int64 `json:"listedPrice,omitempty"`
	ListedAt    *int64 `json:"listedAt,omitempty"`

	BuyTrade *TradeRecord `gorm:"foreignKey:BuyTradeID" json:"buyTrade,omitempty"`
}

func (InventoryItem) TableName() string { return "inventory" }
