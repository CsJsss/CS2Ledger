package youpin

import "encoding/json"

type youpinBuyProduct struct {
	AssertID          int64       `json:"assertId"`
	CommodityID       int64       `json:"commodityId"`
	CommodityName     string      `json:"commodityName"`
	CommodityHashName string      `json:"commodityHashName"` // market hash name, e.g. "M4A4 | Asiimov (Field-Tested)"
	Price             json.Number `json:"price"`             // 元, e.g. 1610.00
	CommodityAmount   json.Number `json:"commodityAmount"`   // 该商品总价 (元)
	CommodityAbrade   string      `json:"commodityAbrade"`   // 磨损度
	ExteriorName      string      `json:"exteriorName"`      // 磨损范围 (e.g., 久经沙场)
	RarityName        string      `json:"rarityName"`        // 稀有度
	ItemSetName       string      `json:"itemSetName"`       // 套装
	TypeName          string      `json:"typeName"`          // 武器类型
	PaintIndex        int         `json:"paintIndex"`
	PaintSeed         int         `json:"paintSeed"`
}

type youpinBuyOrder struct {
	ID              string             `json:"id"`
	OrderID         int64              `json:"orderId"`
	BuyerUserID     int64              `json:"buyerUserId"`
	FinishOrderTime int64              `json:"finishOrderTime"`
	OrderStatusName string             `json:"orderStatusName"`
	CommodityNum    int                `json:"commodityNum"`
	TotalAmount     int64              `json:"totalAmount"` // 订单总价 (分)
	ProductList     []youpinBuyProduct `json:"productDetailList"`
}

type youpinBuyPageResponse struct {
	Code int `json:"code"`
	Data struct {
		OrderList   []youpinBuyOrder `json:"orderList"`
		TotalCount  int              `json:"total"`
		OrderRevert any              `json:"orderRevertInfo"`
	} `json:"data"`
}

type youpinSellOrder struct {
	OrderNo         string      `json:"orderNo"`
	FinishOrderTime int64       `json:"finishOrderTime"`
	PaymentAmount   json.Number `json:"paymentAmount"`  // 收款金额 (元)
	TotalFeeAmount  json.Number `json:"totalFeeAmount"` // 手续费 (元)
	TotalAmount     json.Number `json:"totalAmount"`    // 订单总价 (元)
	CommodityNum    int         `json:"commodityNum"`
	OrderStatusName string      `json:"orderStatusName"`
	ProductDetail   struct {
		AssertID          int64       `json:"assertId"`
		CommodityID       int64       `json:"commodityId"`
		CommodityName     string      `json:"commodityName"`
		CommodityHashName string      `json:"commodityHashName"` // market hash name
		Price             json.Number `json:"price"`             // 单价 (元)
		CommodityAmount   json.Number `json:"commodityAmount"`   // 单品总价 (元)
		CommodityAbrade   string      `json:"commodityAbrade"`   // 磨损度
		ExteriorName      string      `json:"exteriorName"`      // 磨损范围
		RarityName        string      `json:"rarityName"`        // 稀有度
		ItemSetName       string      `json:"itemSetName"`       // 套装
		TypeName          string      `json:"typeName"`          // 武器类型
		PaintIndex        int         `json:"paintIndex"`
		PaintSeed         int         `json:"paintSeed"`
	} `json:"productDetail"`
}

type youpinSellPageResponse struct {
	Code int `json:"code"`
	Data struct {
		OrderList   []youpinSellOrder `json:"orderList"`
		TotalCount  int               `json:"total"`
		OrderRevert any               `json:"orderRevertInfo"`
	} `json:"data"`
}

// --- Balance ---

type youpinBalanceResponse struct {
	Code      int             `json:"code"`
	Msg       string          `json:"msg"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type youpinBalanceData struct {
	Balance                 string                  `json:"balance"`
	BalanceFroze            string                  `json:"balanceFroze"`
	ShowBalance2            bool                    `json:"showBalance2"`
	PreSellTitle            *string                 `json:"preSellTitle"`
	PurchaseBalance         string                  `json:"purchaseBalance"`
	PurchaseBalanceWithdraw string                  `json:"purchaseBalanceWithdraw"`
	PurchaseBalanceTransfer string                  `json:"purchaseBalanceTransfer"`
	AvailableTotalAmountStr string                  `json:"availableTotalAmountStr"`
	AvailableTotalAmount    string                  `json:"availableTotalAmount"`
	TradeOnlyTotalAmountStr string                  `json:"tradeOnlyTotalAmountStr"`
	TradeOnlyTotalAmount    string                  `json:"tradeOnlyTotalAmount"`
	FrozeTotalAmountStr     string                  `json:"frozeTotalAmountStr"`
	FrozeTotalAmount        string                  `json:"frozeTotalAmount"`
	Currency                string                  `json:"currency"`
	ForceUpdate             bool                    `json:"forceUpdate"`
	ShowAccountV2           bool                    `json:"showAccountV2"`
	WithdrawMaxMoneyLabel   *string                 `json:"withdrawMaxMoneyLabel"`
	WithdrawMaxMoney        *string                 `json:"withdrawMaxMoney"`
	ShowBankCardWithdraw    *string                 `json:"showBankCardWithdraw"`
	WithdrawInfo            *string                 `json:"withdrawInfo"`
	TipContent              string                  `json:"tipContent"`
	List                    []youpinBalanceListItem `json:"list"`
}

type youpinBalanceListItem struct {
	BalanceTitle       string `json:"balanceTitle"`
	Amount             string `json:"amount"`
	AmountStr          string `json:"amountStr"`
	AvailableAmount    string `json:"availableAmount"`
	AvailableAmountStr string `json:"availableAmountStr"`
	FrozeAmount        string `json:"frozeAmount"`
	FrozeAmountStr     string `json:"frozeAmountStr"`
	WithdrawTitle      string `json:"withdrawTitle"`
	QuestionDesc       string `json:"questionDesc"`
	Type               int    `json:"type"`
	BalanceType        int    `json:"balanceType"`
}

// --- Bill / Fund Flow ---

type youpinBillItem struct {
	TypeID    int    `json:"typeId"`
	TypeName  string `json:"typeName"`
	ThisMoney string `json:"thisMoney"` // 元, e.g. "-426.00"
	OrderNo   string `json:"orderNo"`
	AddTime   string `json:"addTime"` // "2026-05-17 11:49:30"
}

type youpinBillPageResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Total    int              `json:"total"`
		DataList []youpinBillItem `json:"dataList"`
	} `json:"data"`
}
