package model

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	Name              string `gorm:"not null;uniqueIndex" json:"name"`
	Platform          string `gorm:"not null" json:"platform"`
	Cookie            string `gorm:"not null" json:"-"`
	AvailableBalance  int64  `gorm:"not null;default:0" json:"availableBalance"`
	PurchaseBalance   int64  `gorm:"not null;default:0" json:"purchaseBalance"`
	WithdrawalFeeRate int64  `gorm:"not null;default:0" json:"withdrawalFeeRate"`
	Remark            string `json:"remark"`
	Status            string `gorm:"not null;default:active" json:"status"`
	LastSyncAt        *int64 `json:"lastSyncAt"`
}

func (Account) TableName() string { return "accounts" }
