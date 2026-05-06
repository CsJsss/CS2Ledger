package model

import "gorm.io/gorm"

type InventoryItem struct {
	gorm.Model
	AccountID   uint   `gorm:"not null;uniqueIndex:idx_inventory_asset" json:"accountId"`
	AssetID     string `gorm:"not null;uniqueIndex:idx_inventory_asset" json:"assetId"`
	ItemName    string `gorm:"not null" json:"itemName"`
	Exterior    string `json:"exterior"`
	PaintSeed   *int64 `json:"paintSeed"`
	BuyTradeID  uint   `gorm:"not null" json:"buyTradeId"`
	Status      string `gorm:"not null;default:in_inventory;index:idx_inventory_status" json:"status"`
	ListedPrice *int64 `json:"listedPrice"`
	ListedAt    *int64 `json:"listedAt"`

	BuyTrade *TradeRecord `gorm:"foreignKey:BuyTradeID" json:"-"`
}

func (InventoryItem) TableName() string { return "inventory" }
