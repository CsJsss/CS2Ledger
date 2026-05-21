package buff

// --- Shared ---

// buffResponse is a generic envelope for BUFF API responses.
type buffResponse[T any] struct {
	Code string  `json:"code"`
	Msg  *string `json:"msg,omitempty"`
	Data T       `json:"data"`
}

// tradeHistoryData is shared by buy/sell order history responses.
type tradeHistoryData[T any] struct {
	Items      []T                 `json:"items"`
	GoodsInfos map[string]goodInfo `json:"goods_infos"`
	TotalPages int                 `json:"total_page"`
	Total      int                 `json:"total_count"`
}

type goodInfo struct {
	Name           string   `json:"name"`
	ShortName      string   `json:"short_name"`
	MarketHashName string   `json:"market_hash_name"`
	IconURL        string   `json:"icon_url"`
	SteamPriceCny  string   `json:"steam_price_cny"`
	SellMinPrice   string   `json:"sell_min_price"`
	Tags           goodTags `json:"tags"`
}

type goodTags struct {
	Exterior      tagItem `json:"exterior"`
	Rarity        tagItem `json:"rarity"`
	Category      tagItem `json:"category"`
	CategoryGroup tagItem `json:"category_group"`
	Type          tagItem `json:"type"`
	Quality       tagItem `json:"quality"`
	Series        tagItem `json:"series"`
	Itemset       tagItem `json:"itemset"`
	WeaponCase    tagItem `json:"weaponcase"`
	Custom        tagItem `json:"custom"`
}

type tagItem struct {
	LocalizedName string `json:"localized_name"`
}

type stickerInfo struct {
	StickerID      int      `json:"sticker_id"`
	Slot           int      `json:"slot"`
	Wear           float64  `json:"wear"`
	Name           string   `json:"name"`
	ImageURL       string   `json:"img_url"`
	ReferencePrice string   `json:"sell_reference_price"`
	OffsetX        *float64 `json:"offset_x"`
	OffsetY        *float64 `json:"offset_y"`
}

type keychainInfo struct {
	Name     string `json:"name"`
	ImageURL string `json:"img_url"`
}

type assetInfo struct {
	AssetID              string          `json:"assetid"`
	ClassID              string          `json:"classid"`
	InstanceID           string          `json:"instanceid"`
	PaintWear            string          `json:"paintwear"`
	TradableCooldownText string          `json:"tradable_cooldown_text"`
	TradableUnfrozenTime *int64          `json:"tradable_unfrozen_time"`
	HasTradableCooldown  bool            `json:"has_tradable_cooldown"`
	Info                 assetInfoDetail `json:"info"`
}

type assetInfoDetail struct {
	PaintSeed       int            `json:"paintseed"`
	PaintIndex      int            `json:"paintindex"`
	Stickers        []stickerInfo  `json:"stickers"`
	Keychains       []keychainInfo `json:"keychains"`
	IconURL         string         `json:"icon_url"`
	OriginalIconURL string         `json:"original_icon_url"`
	FraudWarnings   string         `json:"fraudwarnings"`
}

// --- Verify ---

type userInfoResponse struct {
	Code string `json:"code"`
}

// --- Buy history ---

type buyOrderItem struct {
	ID           string    `json:"id"`
	State        string    `json:"state"`
	StateText    string    `json:"state_text"`
	TransactTime int64     `json:"transact_time"`
	CreatedAt    int64     `json:"created_at"`
	Price        string    `json:"price"`
	Income       string    `json:"income"`
	Fee          string    `json:"fee"`
	GoodsID      int64     `json:"goods_id"`
	AssetInfo    assetInfo `json:"asset_info"`
	TradeOfferID string    `json:"tradeofferid"`
	BuyerPayTime int64     `json:"buyer_pay_time"`
}

// --- Sell history ---

type sellOrderItem struct {
	ID           string    `json:"id"`
	GoodsID      int64     `json:"goods_id"`
	Price        string    `json:"price"`
	Fee          string    `json:"fee"`
	Income       string    `json:"income"`
	State        string    `json:"state"`
	StateText    string    `json:"state_text"`
	CreateTime   int64     `json:"created_at"`
	TransactTime int64     `json:"transact_time"`
	AssetInfo    assetInfo `json:"asset_info"`
	TradeOfferID string    `json:"tradeofferid"`
}

// --- Balance ---

type balanceData struct {
	CashAmount                     string               `json:"cash_amount"`
	CashAmountOuter                string               `json:"cash_amount_outer"`
	CashAmountInner                string               `json:"cash_amount_inner"`
	SecurityAmount                 string               `json:"security_amount"`
	FrozenAmount                   string               `json:"frozen_amount"`
	EpayAmount                     string               `json:"epay_amount"`
	EpayAbleWithdrawAmount         string               `json:"epay_able_withdraw_amount"`
	EpayUnableWithdrawAmount       string               `json:"epay_unable_withdraw_amount"`
	EpayFrozenAmount               string               `json:"epay_frozen_amount"`
	AlipayAmount                   string               `json:"alipay_amount"`
	AlipayAbleWithdrawAmount       string               `json:"alipay_able_withdraw_amount"`
	AlipayUnableWithdrawAmount     string               `json:"alipay_unable_withdraw_amount"`
	AlipayFrozenAmount             string               `json:"alipay_frozen_amount"`
	AlipayWalletAmount             string               `json:"alipay_wallet_amount"`
	AlipayWalletFrozenAmount       string               `json:"alipay_wallet_frozen_amount"`
	AlipayWalletAbleWithdrawAmount string               `json:"alipay_wallet_able_withdraw_amount"`
	TotalAbleWithdrawAmount        string               `json:"total_able_withdraw_amount"`
	TotalUnableWithdrawAmount      string               `json:"total_unable_withdraw_amount"`
	RechargeShowAmount             string               `json:"recharge_show_amount"`
	PendingDivideAmount            string               `json:"pending_divide_amount"`
	Realname                       bool                 `json:"realname"`
	HasAdminFrozenAsset            bool                 `json:"has_admin_frozen_asset"`
	AllowLargeAmountWithdraw       bool                 `json:"allow_large_amount_withdraw"`
	AllowLargeAmountWithdrawNew    bool                 `json:"allow_large_amount_withdraw_new"`
	RemainWithdrawCounts           remainWithdrawCounts `json:"remain_withdraw_counts"`
}

type remainWithdrawCounts struct {
	Epay         int `json:"epay"`
	Alipay       int `json:"alipay"`
	AlipayWallet int `json:"alipay_wallet"`
	Together     int `json:"together"`
	Airwallex    int `json:"airwallex"`
	Payoneer     int `json:"payoneer"`
}
