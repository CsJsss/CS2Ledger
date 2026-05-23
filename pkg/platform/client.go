package platform

import (
	"context"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

const (
	PlatformBuff   = "buff"
	PlatformYoupin = "youpin"
	PlatformC5     = "c5"
	PlatformIGXE   = "igxe"
	PlatformECO    = "eco"
	PlatformCsqaq  = "csqaq"
)

// TradeState controls whether history queries filter by order completion status.
type TradeState string

const (
	TradeStateAll       TradeState = ""          // no filter, return all orders
	TradeStateCompleted TradeState = "completed" // only completed / successful orders
)

// TradeRecord is a unified trade model returned by all platform clients.
// UnitPrice and TotalPrice are in cents. TradeAt is a unix millisecond timestamp.
type TradeRecord struct {
	model.CS2Item
	ExternalID   string
	TradeType    string
	Quantity     int64
	UnitPrice    int64
	TotalPrice   int64
	Fee          int64
	TradeAt      int64
	State        string
	StateText    string
	TransactTime int64
	TradeOfferID string
}

// Balance holds account balance in yuan.
type Balance struct {
	Available float64 // 钱包余额
	Purchase  float64 // 求购余额
	Frozen    float64 // 冻结余额
	Instant   float64 // 秒到账余额
}

// QueryConfig holds optional parameters for history queries.
type QueryConfig struct {
	Since       int64             // unix ms, 0 = no filter
	Limit       int               // max records, 0 = no limit
	TradeState  TradeState        // order completion filter, default = all
	ExtraParams map[string]string // merged into HTTP request params
}

// QueryOption is a functional option for history query methods.
type QueryOption func(*QueryConfig)

// WithSince filters records with TradeAt >= since (unix ms).
func WithSince(since int64) QueryOption {
	return func(c *QueryConfig) { c.Since = since }
}

// WithExtraParams adds extra key-value pairs to the HTTP request parameters.
func WithExtraParams(params map[string]string) QueryOption {
	return func(c *QueryConfig) {
		c.ExtraParams = params
	}
}

// WithTradeState controls whether to filter by order completion status.
func WithTradeState(s TradeState) QueryOption {
	return func(c *QueryConfig) { c.TradeState = s }
}

// WithLimit caps the number of records returned.
func WithLimit(limit int) QueryOption {
	return func(c *QueryConfig) { c.Limit = limit }
}

func ApplyQueryOpts(opts []QueryOption) QueryConfig {
	cfg := QueryConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// Client is the interface all platform clients must implement.
type Client interface {
	Verify(ctx context.Context) error
	GetBuyHistory(ctx context.Context, opts ...QueryOption) ([]TradeRecord, error)
	GetSellHistory(ctx context.Context, opts ...QueryOption) ([]TradeRecord, error)
	GetBalance(ctx context.Context) (*Balance, error)
}
