package platform

import "context"

type PriceInfo struct {
	MarketHashName string  `json:"marketHashName"`
	BuffPrice      float64 `json:"buffPrice"`
	BuffVolume     int     `json:"buffVolume"`
	YoupinPrice    float64 `json:"youpinPrice"`
	YoupinVolume   int     `json:"youpinVolume"`
	SteamPrice     float64 `json:"steamPrice"`
	SteamVolume    int     `json:"steamVolume"`
}

type PriceProvider interface {
	GetPrices(ctx context.Context, marketHashNames []string) ([]PriceInfo, error)
	ResolveGoodsInfo(ctx context.Context, itemName, exterior string) (goodID int, marketHashName string, err error)
}
