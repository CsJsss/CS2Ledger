package buff

// --- Shared ---

type goodInfo struct {
	ShortName string `json:"short_name"`
}

// --- Verify ---

type userInfoResponse struct {
	Code string `json:"code"`
}

// --- Buy history ---

type buyOrderHistoryResponse struct {
	Code string              `json:"code"`
	Data buyOrderHistoryData `json:"data"`
}

type buyOrderHistoryData struct {
	Items      []buyOrderItem      `json:"items"`
	GoodsInfos map[string]goodInfo `json:"goods_infos"`
	TotalPages int                 `json:"total_pages"`
	Total      int                 `json:"total"`
	PageSize   int                 `json:"page_size"`
}

type buyOrderItem struct {
	State        string            `json:"state"`
	TransactTime int64             `json:"transact_time"`
	Income       string            `json:"income"`
	GoodsID      int64             `json:"goods_id"`
	AssetInfo    buyOrderAssetInfo `json:"asset_info"`
}

type buyOrderAssetInfo struct {
	AssetID string `json:"assetid"`
}

// --- Sell history ---

type sellOrderHistoryResponse struct {
	Code string               `json:"code"`
	Data sellOrderHistoryData `json:"data"`
}

type sellOrderHistoryData struct {
	Items      []sellOrderItem     `json:"items"`
	GoodsInfos map[string]goodInfo `json:"goods_infos"`
	TotalPages int                 `json:"total_pages"`
	Total      int                 `json:"total"`
	PageSize   int                 `json:"page_size"`
}

type sellOrderItem struct {
	ID         string `json:"id"`
	GoodsID    int64  `json:"goods_id"`
	Price      string `json:"price"`
	Status     string `json:"status"`
	CreateTime int64  `json:"create_time"`
}

// --- Balance ---

type balanceResponse struct {
	Code string      `json:"code"`
	Data balanceData `json:"data"`
}

type balanceData struct {
	Balance       string `json:"balance"`
	FrozenBalance string `json:"frozen_balance"`
}
