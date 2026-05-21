package c5

// Order status constants. Two numbering schemes exist:
// v1/detail endpoints use 0-11; v2 list endpoints use 2xx.
const (
	StatusPendingPay     = 0   // 待付款
	StatusPendingDeliver = 1   // 待发货
	StatusDelivering     = 2   // 发货中
	StatusPendingReceive = 3   // 待收货
	StatusCompleted      = 10  // 已完成 (v1 / detail)
	StatusCancelled      = 11  // 已取消
	StatusSuccessV2      = 200 // 已完成 (v2 list)
)

func isCompletedStatus(s int) bool {
	return s == StatusCompleted || s == StatusSuccessV2
}

// c5Response is a generic envelope for C5 API responses.
type c5Response[T any] struct {
	Success      bool   `json:"success"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMsg     string `json:"errorMsg"`
	ErrorCodeStr string `json:"errorCodeStr"`
	Data         T      `json:"data"`
}

// Balance v2

type c5BalanceV2Data struct {
	UserID            string  `json:"userId"`
	MoneyAmount       float64 `json:"moneyAmount"`
	DepositAmount     float64 `json:"depositAmount"`
	TradeSettleAmount float64 `json:"tradeSettleAmount"`
	CreditMoney       float64 `json:"creditMoney"`
	CreditDeposit     float64 `json:"creditDeposit"`
}

// Buyer order v2 (POST /merchant/order/v2/buyer/status)

type c5BuyerOrderData struct {
	Total string         `json:"total"`
	Pages int            `json:"pages"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	List  []c5BuyerOrder `json:"list"`
}

type c5BuyerOrder struct {
	OrderID        string  `json:"orderId"`
	ProductID      string  `json:"productId"`
	Price          float64 `json:"price"`
	BuyerFee       float64 `json:"buyerFee"`
	StatusName     string  `json:"statusName"`
	Status         int     `json:"status"`
	DeliverType    int     `json:"deliverType"`
	ReceiveSteamID string  `json:"receiveSteamId"`
	CreateTime     int64   `json:"createTime"`
	Type           int     `json:"type"`
	TradeOfferID   *string `json:"tradeOfferId"`
}

// Seller order v1 (GET /merchant/order/v1/list)

type c5SellerOrderData struct {
	Total string          `json:"total"`
	Pages int             `json:"pages"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	List  []c5SellerOrder `json:"list"`
}

type c5SellerOrder struct {
	OrderID          string              `json:"orderId"`
	ProductID        string              `json:"productId"`
	Name             string              `json:"name"`
	MarketHashName   string              `json:"marketHashName"`
	Price            float64             `json:"price"`
	Status           int                 `json:"status"`
	OrderConfirmInfo *c5OrderConfirmInfo `json:"orderConfirmInfoDTO"`
}

type c5OrderConfirmInfo struct {
	OrderCreateTime int64  `json:"orderCreateTime"`
	StatusName      string `json:"statusName"`
}

// Order detail v2 (GET /merchant/order/v2/buy/detail)

type c5BuyerOrderDetail struct {
	OrderID        string          `json:"orderId"`
	ProductID      string          `json:"productId"`
	Price          float64         `json:"price"`
	Status         int             `json:"status"`
	StatusName     string          `json:"statusName"`
	CreateTime     int64           `json:"createTime"`
	DeliverType    int             `json:"deliverType"`
	ReceiveSteamID string          `json:"receiveSteamId"`
	NewAssetID     string          `json:"newAssetId"`
	OpenItemInfo   *c5OpenItemInfo `json:"openItemInfo"`
	OfferInfoDTO   *c5OfferInfo    `json:"offerInfoDTO"`
	AssetInfo      *c5AssetInfo    `json:"assetInfo"`
}

type c5AssetInfo struct {
	AssetID    string `json:"assetId"`
	InstanceID string `json:"instanceId"`
	Wear       string `json:"wear"`
	PaintIndex int    `json:"paintIndex"`
	PaintSeed  int    `json:"paintSeed"`
}

type c5OpenItemInfo struct {
	Name           string `json:"name"`
	MarketHashName string `json:"marketHashName"`
	AppID          int    `json:"appId"`
	ItemID         string `json:"itemId"`
}

type c5OfferInfo struct {
	TradeOfferID string `json:"tradeOfferId"`
}
