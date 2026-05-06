package youpin

type youpinBuyProduct struct {
	AssertID      int64  `json:"assertId,string"`
	CommodityID   int64  `json:"commodityId,string"`
	CommodityName string `json:"commodityName"`
	Price         int64  `json:"price,string"`
	Abrade        string `json:"abrade"`
	TypeName      string `json:"typeName"`
}

type youpinBuyOrder struct {
	ID              int64              `json:"id,string"`
	OrderID         int64              `json:"orderId,string"`
	BuyerUserID     int64              `json:"buyerUserId,string"`
	FinishOrderTime int64              `json:"finishOrderTime,string"`
	OrderStatusName string             `json:"orderStatusName"`
	CommodityNum    int                `json:"commodityNum,string"`
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
	OrderNo         string `json:"orderNo"`
	FinishOrderTime int64  `json:"finishOrderTime,string"`
	PaymentAmount   int64  `json:"paymentAmount,string"`
	CommodityNum    int    `json:"commodityNum,string"`
	ProductDetail   struct {
		AssertID      int64  `json:"assertId,string"`
		CommodityName string `json:"commodityName"`
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
