package model

import "gorm.io/gorm"

// BillType constants are our internal classification for common transaction types.
// Each platform converts its own type_id into one of these constants via a mapping
// function (e.g. youpinTypeToInternal). Unrecognized types fall back to BillTypeOther.
//
// TypeID vs TypeName:
//   - TypeID is our internal constant (BillTypePurchase, etc.). Frontend uses it for
//     color coding and filtering across platforms.
//   - TypeName is a unified label from BillTypeName(). When TypeID == BillTypeOther,
//     the converter falls back to the platform's original type name.
const (
	BillTypePurchase                  = 1  // 购买
	BillTypeSell                      = 2  // 出售
	BillTypeRentalIncome              = 3  // 收取租金
	BillTypeRenewalRental             = 4  // 收取续租资金
	BillTypeRentalFee                 = 5  // 租赁服务费
	BillTypeRecharge                  = 6  // 充值
	BillTypeWithdraw                  = 7  // 提现
	BillTypeRefund                    = 8  // 退款
	BillTypeRechargForPurchaseAccount = 9  // 求购账户充值
	BillTypeWithdrawRefund            = 10 // 提现退款
	BillTypeOther                     = 99 // 其他 — 回退到平台原始 TypeName 展示
)

var billTypeNames = map[int]string{
	BillTypePurchase:                  "购买",
	BillTypeSell:                      "出售",
	BillTypeRentalIncome:              "收取租金",
	BillTypeRenewalRental:             "收取续租资金",
	BillTypeRentalFee:                 "租赁服务费",
	BillTypeRecharge:                  "充值",
	BillTypeWithdraw:                  "提现",
	BillTypeRefund:                    "退款",
	BillTypeRechargForPurchaseAccount: "求购账户充值",
	BillTypeWithdrawRefund:            "提现退款",
}

func BillTypeName(t int) string {
	if name, ok := billTypeNames[t]; ok {
		return name
	}
	return ""
}

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
