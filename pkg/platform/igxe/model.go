package igxe

type totalMoneyResponse struct {
	ResultCode string `json:"ResultCode"`
	ResultData struct {
		Money    float64 `json:"Money"`
		UserName string  `json:"UserName"`
	} `json:"ResultData"`
}

type sellerOrderListResponse struct {
	ResultCode string              `json:"ResultCode"`
	ResultMsg  string              `json:"ResultMsg"`
	ResultData sellerOrderListData `json:"ResultData"`
}

type sellerOrderListData struct {
	PageResult []igxeOrder `json:"PageResult"`
}

type igxeOrder struct {
	OrderNum   string  `json:"OrderNum"`
	GoodsName  string  `json:"GoodsName"`
	Price      float64 `json:"Price"`
	State      int     `json:"State"`
	CreateDate string  `json:"CreateDate"`
	AssetID    string  `json:"AssetId"`
}
