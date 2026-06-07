package platform

//go:generate go tool mockery

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

// BillRecord is a unified bill/transaction record returned by platform clients.
// ThisMoney is in cents; positive = income, negative = expense.
// AddTime is a unix millisecond timestamp.
type BillRecord struct {
	TypeName  string
	TypeID    int
	ThisMoney int64
	OrderNo   string
	AddTime   int64
}

// QueryConfig holds optional parameters for history queries.
type QueryConfig struct {
	Since       int64             // unix ms, 0 = no filter
	Limit       int               // max records, 0 = no limit
	TradeState  TradeState        // order completion filter, default = all
	Page        int               // single page to fetch, 0 = all pages
	PageSize    int               // page size, 0 = platform default
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

// WithPage fetches a single page (1-based). 0 = fetch all pages.
func WithPage(page int) QueryOption {
	return func(c *QueryConfig) { c.Page = page }
}

// WithPageSize sets the page size for paginated requests.
func WithPageSize(pageSize int) QueryOption {
	return func(c *QueryConfig) { c.PageSize = pageSize }
}

func ApplyQueryOpts(opts []QueryOption) QueryConfig {
	cfg := QueryConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// Client is the interface all platform clients must implement.
//
//mockery:generate: true
//mockery:filename: client_mock_test.go
type Client interface {
	Verify(ctx context.Context) error
	GetBuyHistory(ctx context.Context, opts ...QueryOption) ([]TradeRecord, error)
	GetSellHistory(ctx context.Context, opts ...QueryOption) ([]TradeRecord, error)
	GetBalance(ctx context.Context) (*Balance, error)
	GetBillHistory(ctx context.Context, opts ...QueryOption) ([]BillRecord, error)
}
