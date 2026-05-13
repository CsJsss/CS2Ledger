package youpin

import "encoding/json"

type youpinBuyProduct struct {
	AssertID        int64       `json:"assertId"`
	CommodityID     int64       `json:"commodityId"`
	CommodityName   string      `json:"commodityName"`
	Price           json.Number `json:"price"`           // 元, e.g. 1610.00
	CommodityAmount json.Number `json:"commodityAmount"` // 该商品总价 (元)
	CommodityAbrade string      `json:"commodityAbrade"` // 磨损度
	ExteriorName    string      `json:"exteriorName"`    // 磨损范围 (e.g., 久经沙场)
	RarityName      string      `json:"rarityName"`      // 稀有度
	ItemSetName     string      `json:"itemSetName"`     // 套装
	TypeName        string      `json:"typeName"`        // 武器类型
	PaintIndex      int         `json:"paintIndex"`
	PaintSeed       int         `json:"paintSeed"`
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
		AssertID        int64       `json:"assertId"`
		CommodityID     int64       `json:"commodityId"`
		CommodityName   string      `json:"commodityName"`
		Price           json.Number `json:"price"`           // 单价 (元)
		CommodityAmount json.Number `json:"commodityAmount"` // 单品总价 (元)
		CommodityAbrade string      `json:"commodityAbrade"` // 磨损度
		ExteriorName    string      `json:"exteriorName"`    // 磨损范围
		RarityName      string      `json:"rarityName"`      // 稀有度
		ItemSetName     string      `json:"itemSetName"`     // 套装
		TypeName        string      `json:"typeName"`        // 武器类型
		PaintIndex      int         `json:"paintIndex"`
		PaintSeed       int         `json:"paintSeed"`
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
