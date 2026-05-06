package c5

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils"
)

type Client struct {
	platform.BaseClient
	apiKey string
}

func New(apiKey string) *Client {
	return &Client{
		BaseClient: platform.NewBaseClient(utils.PlatformC5, "http://openapi.c5game.com"),
		apiKey:     apiKey,
	}
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("c5: verifying")
	resp, err := c.doGet(ctx, "/merchant/account/v1/balance", nil)
	if err != nil {
		c.Log.Warn("c5: verify failed", "err", err)
		return fmt.Errorf("c5 verify: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result c5BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("c5 verify: %w", err)
	}
	if !result.Success {
		c.Log.Warn("c5: verify invalid credential")
		return fmt.Errorf("c5 verify: credential invalid")
	}
	c.Log.Info("c5: verify ok")
	return nil
}

func (c *Client) FetchBalance(ctx context.Context) (*platform.Balance, error) {
	c.Log.Info("c5: fetching balance")
	resp, err := c.doGet(ctx, "/merchant/account/v1/balance", nil)
	if err != nil {
		return nil, fmt.Errorf("c5 balance: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result c5BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("c5 balance: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("c5 balance: API error")
	}

	return &platform.Balance{
		Available: int64(result.Data.Amount * 100),
		Purchase:  0,
	}, nil
}

func (c *Client) FetchBuyHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Info("c5: fetching buy history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "buy", 1*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchBuyPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("c5: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	// Status 3 = 待收货 (awaiting receipt), the buyer's completed state.
	params := map[string]string{
		"status": "3",
		"page":   fmt.Sprintf("%d", page),
	}
	resp, err := c.doGet(ctx, "/merchant/order/v1/list", params)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result c5SellListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, err
	}
	if !result.Success {
		return nil, false, fmt.Errorf("c5 API error")
	}

	sinceSec := since / 1000
	trades := make([]platform.TradeRecord, 0, len(result.Data.List))
	for _, item := range result.Data.List {
		if item.CreateTime < sinceSec {
			continue
		}
		price := int64(item.Amount * 100)
		qty := int64(item.CommodityNum)
		if qty == 0 {
			qty = 1
		}
		trades = append(trades, platform.TradeRecord{
			ExternalID: fmt.Sprintf("c5-buy-%s", item.OrderNo),
			AssetID:    item.CommodityID,
			ItemName:   item.CommodityName,
			TradeType:  "buy",
			Quantity:   qty,
			UnitPrice:  price / qty,
			TotalPrice: price,
			Fee:        0,
			TradeAt:    item.CreateTime * 1000,
		})
	}

	hasMore := len(result.Data.List) >= result.Data.Limit
	return trades, hasMore, nil
}

func (c *Client) FetchSellHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Info("c5: fetching sell history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "sell", 1*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("c5: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	params := map[string]string{
		"status": "10",
		"page":   fmt.Sprintf("%d", page),
	}
	resp, err := c.doGet(ctx, "/merchant/order/v1/list", params)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result c5SellListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, err
	}
	if !result.Success {
		return nil, false, fmt.Errorf("c5 API error")
	}

	sinceSec := since / 1000
	trades := make([]platform.TradeRecord, 0, len(result.Data.List))
	for _, item := range result.Data.List {
		if item.CreateTime < sinceSec {
			continue
		}
		price := int64(item.Amount * 100)
		qty := int64(item.CommodityNum)
		if qty == 0 {
			qty = 1
		}
		trades = append(trades, platform.TradeRecord{
			ExternalID: fmt.Sprintf("c5-sell-%s", item.OrderNo),
			AssetID:    item.CommodityID,
			ItemName:   item.CommodityName,
			TradeType:  "sell",
			Quantity:   qty,
			UnitPrice:  price / qty,
			TotalPrice: price,
			Fee:        0,
			TradeAt:    item.CreateTime * 1000,
		})
	}

	hasMore := len(result.Data.List) >= result.Data.Limit
	return trades, hasMore, nil
}

func (c *Client) doGet(ctx context.Context, path string, params map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("app-key", c.apiKey)

	return c.HTTP.Do(req)
}
