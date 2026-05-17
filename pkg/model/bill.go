package model

import "gorm.io/gorm"

const (
	BillTypePurchase     = 1
	BillTypeSell         = 2
	BillTypeRentalIncome = 3
	BillTypeRentalFee    = 4
	BillTypeRecharge     = 5
	BillTypeWithdraw     = 6
	BillTypeRefund       = 7
	BillTypeOther        = 99
)

type BillRecord struct {
	gorm.Model
	AccountID uint   `gorm:"not null;index:idx_bill_account" json:"accountId"`
	Platform  string `gorm:"not null" json:"platform"`
	TypeID    int    `gorm:"not null" json:"typeId"`
	TypeName  string `gorm:"not null" json:"typeName"`
	ThisMoney int64  `gorm:"not null" json:"thisMoney"`
	OrderNo   string `gorm:"index" json:"orderNo"`
	AddTime   int64  `gorm:"not null;index" json:"addTime"`
}

func (BillRecord) TableName() string { return "bill_records" }
