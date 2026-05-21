package igxe

// igxeResponse is a generic envelope for IGXE API responses.
type igxeResponse[T any] struct {
	ResultCode string `json:"ResultCode"`
	ResultMsg  string `json:"ResultMsg"`
	ResultData T      `json:"ResultData"`
}

type totalMoneyData struct {
	Money    float64 `json:"Money"`
	UserName string  `json:"UserName"`
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
