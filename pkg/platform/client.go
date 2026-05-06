package platform

import "context"

// TradeRecord is a unified trade model returned by all platform clients.
// UnitPrice and TotalPrice are in cents. TradeAt is a unix millisecond timestamp.
type TradeRecord struct {
	ExternalID string
	AssetID    string
	ItemName   string
	TradeType  string
	Quantity   int64
	UnitPrice  int64
	TotalPrice int64
	Fee        int64
	TradeAt    int64
}

// Balance holds account balance in cents.
type Balance struct {
	Available int64
	Purchase  int64
}

// Client is the interface all platform clients must implement.
// since is a unix millisecond timestamp — methods should return
// records with TradeAt >= since.
type Client interface {
	Verify(ctx context.Context) error
	FetchBuyHistory(ctx context.Context, since int64) ([]TradeRecord, error)
	FetchSellHistory(ctx context.Context, since int64) ([]TradeRecord, error)
	FetchBalance(ctx context.Context) (*Balance, error)
}
