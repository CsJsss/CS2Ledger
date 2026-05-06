package model

import "gorm.io/gorm"

type PnlDaily struct {
	gorm.Model
	AccountID  uint   `gorm:"not null;uniqueIndex:idx_pnl_daily_account_date" json:"accountId"`
	Date       string `gorm:"not null;uniqueIndex:idx_pnl_daily_account_date" json:"date"`
	TradeCount int64  `gorm:"not null;default:0" json:"tradeCount"`
	GrossPl    int64  `gorm:"not null;default:0" json:"grossPl"`
	Fee        int64  `gorm:"not null;default:0" json:"fee"`
	NetPl      int64  `gorm:"not null;default:0" json:"netPl"`
}

func (PnlDaily) TableName() string { return "pnl_daily" }
