package eco

// Order state codes returned by the ECO API.
const (
	OrderStatePending    = 1  // 等待发货
	OrderStateConfirming = 2  // 等待对方确认
	OrderStateCancelled  = 3  // 交易取消
	OrderStateSuccess    = 6  // 交易成功
	OrderStatePendingPay = 7  // 待付款
	OrderStateShipping   = 8  // 发货中
	OrderStateSuspended  = 9  // 交易暂挂
	OrderStateProtected  = 30 // 交易保护
)

// Trade type codes returned by the ECO API.
const (
	TradeTypeBuy          = 1  // 购买
	TradeTypeGift         = 2  // 赠送
	TradeTypePurchase     = 3  // 求购
	TradeTypeBargain      = 4  // 还价
	TradeTypeRent         = 5  // 租赁
	TradeTypePresale      = 6  // 预售
	TradeTypeResell       = 7  // 转卖
	TradeTypePresaleSell  = 8  // 预售转卖
	TradeTypeFurnaceBuy   = 9  // 合炉购买
	TradeTypeFurnaceGift  = 10 // 合炉赠送
	TradeTypeHourlyRent   = 11 // 小时租
	TradeTypeRentTransfer = 12 // 租赁过户
	TradeTypeBoxOpen      = 13 // 在线开箱
)

// --- Common ---

type ecoResponse[T any] struct {
	ResultCode string `json:"ResultCode"`
	ResultMsg  string `json:"ResultMsg"`
	ResultData T      `json:"ResultData"`
}

// --- Balance ---

type merchantMoneyModel struct {
	UserName            string  `json:"UserName"`
	Money               float64 `json:"Money"`
	LockMoney           float64 `json:"LockMoney"`
	PurchaseMoney       float64 `json:"PurchaseMoney"`
	PurchaseFrozenMoney float64 `json:"PurchaseFrozenMoney"`
}

// --- Buy Orders ---

type buyerOrderModel struct {
	OrderNum         string  `json:"OrderNum"`
	OrderAmount      float64 `json:"OrderAmount"`
	TradeType        string  `json:"TradeType"`
	TradeTypeCode    int     `json:"TradeTypeCode"`
	GoodsName        string  `json:"GoodsName"`
	HashName         string  `json:"HashName"`
	GoodsTotal       int     `json:"GoodsTotal"`
	CreateOrderTime  string  `json:"CreateOrderTime"`
	UpdateTime       string  `json:"UpdateTime"`
	OrderStatus      string  `json:"OrderStatus"`
	OrderStateCode   int     `json:"OrderStateCode"`
	DetailsState     string  `json:"DetailsState"`
	DetailsStateCode int     `json:"DetailsStateCode"`
	CancelReason     string  `json:"CancelReason"`
	Responsible      string  `json:"Responsible"`
	MerchantNo       string  `json:"MerchantNo"`
	PurchaseId       string  `json:"PurchaseId"`
	CdTime           string  `json:"CdTime"`
	TradeTime        string  `json:"TradeTime"`
	SettlementTime   string  `json:"SettlementTime"`
}

type buyerOrderPagesModel struct {
	PageIndex   int               `json:"PageIndex"`
	PageSize    int               `json:"PageSize"`
	TotalRecord int               `json:"TotalRecord"`
	PageResult  []buyerOrderModel `json:"PageResult"`
}

// --- Sell Orders ---

type sellerOrderModel struct {
	OrderNum         string  `json:"OrderNum"`
	OrderAmount      float64 `json:"OrderAmount"`
	TradeType        string  `json:"TradeType"`
	TradeTypeCode    int     `json:"TradeTypeCode"`
	GoodsName        string  `json:"GoodsName"`
	HashName         string  `json:"HashName"`
	GoodsTotal       int     `json:"GoodsTotal"`
	CreateOrderTime  string  `json:"CreateOrderTime"`
	UpdateTime       string  `json:"UpdateTime"`
	OrderStatus      string  `json:"OrderStatus"`
	OrderStateCode   int     `json:"OrderStateCode"`
	DetailsState     string  `json:"DetailsState"`
	DetailsStateCode int     `json:"DetailsStateCode"`
	CancelReason     string  `json:"CancelReason"`
	Responsible      string  `json:"Responsible"`
	SteamId          string  `json:"SteamId"`
	AssetId          string  `json:"AssetId"`
	GameId           string  `json:"GameId"`
	PaintWear        string  `json:"PaintWear"`
	CdTime           string  `json:"CdTime"`
	TradeTime        string  `json:"TradeTime"`
	SettlementTime   string  `json:"SettlementTime"`
	SendOfferRole    int     `json:"SendOfferRole"`
}

type sellerOrderPagesModel struct {
	PageIndex   int                `json:"PageIndex"`
	PageSize    int                `json:"PageSize"`
	TotalRecord int                `json:"TotalRecord"`
	PageResult  []sellerOrderModel `json:"PageResult"`
}

// --- Asset Preview (from detail API) ---

type assetPreviewModel struct {
	GameId        string               `json:"GameId"`
	HashName      string               `json:"HashName"`
	AssetId       string               `json:"AssetId"`
	GoodsImage    string               `json:"GoodsImage"`
	GoodsName     string               `json:"GoodsName"`
	GoodsId       string               `json:"GoodsId"`
	PaintSeed     int                  `json:"PaintSeed"`
	PaintIndex    int                  `json:"PaintIndex"`
	PaintWear     string               `json:"PaintWear"`
	PaintLabelKey string               `json:"PaintLabelKey"`
	PaintLabel    string               `json:"PaintLabel"`
	Fade          float64              `json:"Fade"`
	Stickers      []stickerInfoModel   `json:"Stickers"`
	Keychains     []stickerInfoModel   `json:"Keychains"`
	Gems          []gemsInfoModel      `json:"Gems"`
	Property      []goodsPropertyModel `json:"Property"`
}

type stickerInfoModel struct {
	Id    int     `json:"Id"`
	Name  string  `json:"Name"`
	Icon  string  `json:"Icon"`
	Slot  int     `json:"Slot"`
	Wear  float64 `json:"Wear"`
	Price float64 `json:"Price"`
}

type gemsInfoModel struct {
	GemsId    string  `json:"GemsId"`
	Name      string  `json:"Name"`
	Type      string  `json:"Type"`
	Attribute string  `json:"Attribute"`
	Icon      string  `json:"Icon"`
	Index     int     `json:"Index"`
	Price     float64 `json:"Price"`
}

type goodsPropertyModel struct {
	Category string `json:"Category"`
	Key      string `json:"Key"`
	Value    string `json:"Value"`
	Color    string `json:"Color"`
}

// --- Order Detail ---

type orderDetailModel struct {
	OrderNum          string            `json:"OrderNum"`
	GoodsNo           string            `json:"GoodsNo"`
	GoodsName         string            `json:"GoodsName"`
	GoodsImg          string            `json:"GoodsImg"`
	AssetId           string            `json:"AssetId"`
	OrderStatus       string            `json:"OrderStatus"`
	OrderStateCode    int               `json:"OrderStateCode"`
	TradeType         string            `json:"TradeType"`
	TradeTypeCode     int               `json:"TradeTypeCode"`
	HashName          string            `json:"HashName"`
	SellingPrice      float64           `json:"SellingPrice"`
	TotalMoney        float64           `json:"TotalMoney"`
	CreateOrderTime   string            `json:"CreateOrderTime"`
	FinishOrderTime   string            `json:"FinishOrderTime"`
	PaySuccessTime    string            `json:"PaySuccessTime"`
	CancelReason      string            `json:"CancelReason"`
	Responsible       string            `json:"Responsible"`
	BuyerFee          float64           `json:"BuyerFee"`
	TradeOfferId      string            `json:"TradeOfferId"`
	TradeOfferStatus  string            `json:"TradeOfferStatus"`
	MerchantNo        string            `json:"MerchantNo"`
	PurchaseId        string            `json:"PurchaseId"`
	CdTime            string            `json:"CdTime"`
	TradeTime         string            `json:"TradeTime"`
	SettlementTime    string            `json:"SettlementTime"`
	AssetPreviewModel assetPreviewModel `json:"AssetPreviewModel"`
}

// --- Fund Flow ---

type fundFlowItemModel struct {
	OrderID     string  `json:"OrderID"`
	Amount      float64 `json:"Amount"`
	Type        string  `json:"Type"`
	LastAmount  float64 `json:"LastAmount"`
	AfterAmount float64 `json:"AfterAmount"`
	CreateTime  string  `json:"CreateTime"`
	FounType    string  `json:"FounType"`
}

type fundFlowPagesModel struct {
	PageIndex   int                 `json:"PageIndex"`
	PageSize    int                 `json:"PageSize"`
	TotalRecord int                 `json:"TotalRecord"`
	PageResult  []fundFlowItemModel `json:"PageResult"`
}
