package c5

type c5Order struct {
	OrderNo       string  `json:"orderNo"`
	CommodityName string  `json:"commodityName"`
	CommodityID   string  `json:"commodityId"`
	UnitPrice     float64 `json:"unitPrice"`
	Amount        float64 `json:"amount"`
	CreateTime    int64   `json:"createTime"`
	CommodityNum  int     `json:"commodityNum"`
}

type c5BalanceResponse struct {
	Success bool          `json:"success"`
	Data    c5BalanceData `json:"data"`
}

type c5BalanceData struct {
	Amount float64 `json:"amount"`
}

type c5SellListResponse struct {
	Success bool           `json:"success"`
	Data    c5SellListData `json:"data"`
}

type c5SellListData struct {
	List  []c5Order `json:"list"`
	Limit int       `json:"limit"`
}
