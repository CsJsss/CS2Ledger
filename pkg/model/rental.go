package model

import "gorm.io/gorm"

type RentalRecord struct {
	gorm.Model
	AccountID    uint   `gorm:"not null;index:idx_rental_account" json:"accountId"`
	AssetID      string `gorm:"not null;index:idx_rental_asset" json:"assetId"`
	ItemName     string `gorm:"not null" json:"itemName"`
	Income       int64  `gorm:"not null" json:"income"`
	DurationDays int64  `gorm:"not null" json:"durationDays"`
	StartAt      int64  `gorm:"not null" json:"startAt"`
	EndAt        int64  `gorm:"not null" json:"endAt"`
	ExternalID   string `gorm:"uniqueIndex:idx_rental_external" json:"externalId"`
}

func (RentalRecord) TableName() string { return "rental_records" }
