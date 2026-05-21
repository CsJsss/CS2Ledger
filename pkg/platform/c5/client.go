package c5

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

const (
	defaultPageSize = "20"
)

type Client struct {
	platform.BaseClient
	apiKey string
}

func New(apiKey string, logger *logfx.Logger) *Client {
	return &Client{
		BaseClient: platform.NewBaseClient(platform.PlatformC5, "https://openapi.c5game.com", logger),
		apiKey:     apiKey,
	}
}

// parseC5 unmarshals a c5Response envelope and checks Success.
func parseC5[T any](body []byte) (T, error) {
	var result c5Response[T]
	if err := json.Unmarshal(body, &result); err != nil {
		var zero T
		return zero, err
	}
	if !result.Success {
		var zero T
		return zero, fmt.Errorf("c5 API error: code=%d msg=%s", result.ErrorCode, result.ErrorMsg)
	}
	return result.Data, nil
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("c5: verifying")
	_, body, err := c.doRequest(ctx, "GET", "/merchant/account/v2/balance", nil, nil)
	if err != nil {
		c.Log.Warn("c5: verify failed", "err", err)
		return fmt.Errorf("c5 verify: %w", err)
	}

	if _, err := parseC5[c5BalanceV2Data](body); err != nil {
		c.Log.Warn("c5: verify invalid credential", "err", err)
		return fmt.Errorf("c5 verify: credential invalid: %w", err)
	}
	c.Log.Info("c5: verify ok")
	return nil
}

func (c *Client) GetBalance(ctx context.Context) (*platform.Balance, error) {
	c.Log.Info("c5: fetching balance")
	_, body, err := c.doRequest(ctx, "GET", "/merchant/account/v2/balance", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("c5 balance: %w", err)
	}

	data, err := parseC5[c5BalanceV2Data](body)
	if err != nil {
		return nil, fmt.Errorf("c5 balance: %w", err)
	}

	return &platform.Balance{
		Available: data.MoneyAmount,
		Frozen:    data.TradeSettleAmount,
		Instant:   data.CreditMoney,
	}, nil
}

func (c *Client) GetBillHistory(_ context.Context, _ ...platform.QueryOption) ([]platform.BillRecord, error) {
	return nil, nil
}

func (c *Client) GetBuyHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("c5: fetching buy history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "buy", 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchBuyPage(ctx, page, cfg.Since, cfg.TradeState)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("c5: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64, tradeState platform.TradeState) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"pageNum":  page,
		"pageSize": defaultPageSize,
	}
	_, respBody, err := c.doRequest(ctx, "POST", "/merchant/order/v2/buyer/status", nil, body)
	if err != nil {
		return nil, false, err
	}

	data, err := parseC5[c5BuyerOrderData](respBody)
	if err != nil {
		return nil, false, err
	}

	sinceSec := since / 1000
	filtered := make([]c5BuyerOrder, 0, len(data.List))
	finished := false
	for _, item := range data.List {
		if item.CreateTime < sinceSec {
			finished = true
			continue
		}
		if tradeState == platform.TradeStateCompleted && !isCompletedStatus(item.Status) {
			continue
		}
		filtered = append(filtered, item)
	}

	trades := c.enrichBuyerOrders(ctx, filtered)

	if len(data.List) == 0 || finished {
		return trades, false, nil
	}
	hasMore := page < data.Pages
	return trades, hasMore, nil
}

func (c *Client) enrichBuyerOrders(ctx context.Context, orders []c5BuyerOrder) []platform.TradeRecord {
	if len(orders) == 0 {
		return nil
	}
	throttle := time.NewTicker(700 * time.Millisecond)
	defer throttle.Stop()

	results := make([]platform.TradeRecord, 0, len(orders))
	for _, order := range orders {
		select {
		case <-ctx.Done():
			return results
		case <-throttle.C:
		}
		detail, err := c.getOrderDetail(ctx, order.OrderID)
		if err != nil {
			c.Log.Warn("c5: order detail failed, using list data", "orderId", order.OrderID, "err", err)
			results = append(results, toBuyerTrade(order))
			continue
		}
		results = append(results, toBuyerTradeEnriched(order, detail))
	}
	return results
}

func (c *Client) getOrderDetail(ctx context.Context, orderID string) (c5BuyerOrderDetail, error) {
	query := map[string]string{"orderId": orderID}
	_, body, err := c.doRequest(ctx, "GET", "/merchant/order/v2/buy/detail", query, nil)
	if err != nil {
		return c5BuyerOrderDetail{}, err
	}
	return parseC5[c5BuyerOrderDetail](body)
}

func (c *Client) GetSellHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("c5: fetching sell history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "sell", 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("c5: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string) ([]platform.TradeRecord, bool, error) {
	query := map[string]string{

		"page":  strconv.Itoa(page),
		"limit": defaultPageSize,
	}
	if tradeState == platform.TradeStateCompleted {
		query["status"] = strconv.Itoa(StatusCompleted)
	}
	for k, v := range extra {
		query[k] = v
	}
	_, body, err := c.doRequest(ctx, "GET", "/merchant/order/v1/list", query, nil)
	if err != nil {
		return nil, false, err
	}

	data, err := parseC5[c5SellerOrderData](body)
	if err != nil {
		return nil, false, err
	}

	trades := make([]platform.TradeRecord, 0, len(data.List))
	finished := false
	for _, item := range data.List {
		tradeAt := int64(0)
		if item.OrderConfirmInfo != nil {
			tradeAt = item.OrderConfirmInfo.OrderCreateTime * 1000
		}
		if since > 0 && tradeAt < since {
			finished = true
			continue
		}
		if tradeState == platform.TradeStateCompleted && !isCompletedStatus(item.Status) {
			continue
		}
		trades = append(trades, toSellerTrade(item))
	}

	if len(data.List) == 0 || finished {
		return trades, false, nil
	}
	hasMore := page < data.Pages
	return trades, hasMore, nil
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("User-Agent", platform.RandomUA())
	h.Set("Content-Type", "application/json")
	return h
}

func (c *Client) doRequest(ctx context.Context, method, path string, query map[string]string, body map[string]any) (int, []byte, error) {
	if query == nil {
		query = make(map[string]string)
	}
	query["app-key"] = c.apiKey

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	return c.DoRequest(ctx, method, path, query, bodyBytes, c.headers())
}
