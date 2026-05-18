package model

import "gorm.io/gorm"

// BillType constants are our internal classification for common transaction types.
// Each platform converts its own type_id into one of these constants via a mapping
// function (e.g. youpinTypeToInternal). Unrecognized types fall back to BillTypeOther.
//
// TypeID vs TypeName:
//   - TypeID is our internal constant (BillTypePurchase, etc.). Frontend uses it for
//     color coding and filtering across platforms.
//   - TypeName is the platform's original type label (e.g. "购买饰品", "提现").
//     When TypeID == BillTypeOther, the frontend displays TypeName as-is since there
//     is no standardized mapping for that type.
const (
	BillTypePurchase                  = 1  // 购买
	BillTypeSell                      = 2  // 出售
	BillTypeRentalIncome              = 3  // 收取租金
	BillTypeRentalFee                 = 4  // 租赁服务费
	BillTypeRecharge                  = 5  // 充值
	BillTypeWithdraw                  = 6  // 提现
	BillTypeRefund                    = 7  // 退款
	BillTypeRechargForPurchaseAccount = 8  // 求购账户充值
	BillTypeOther                     = 99 // 其他 — 回退到平台原始 TypeName 展示
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
