package eco

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

const (
	apiMaxDays = 30
	dateFormat = "2006-01-02 15:04"
)

type Client struct {
	platform.BaseClient
	partnerID  string
	privateKey *rsa.PrivateKey
}

// New accepts credential in format "partnerId:privateKeyPEM".
func New(credential string, logger *logfx.Logger) (*Client, error) {
	idx := strings.Index(credential, ":")
	if idx < 0 {
		return nil, fmt.Errorf("eco: credential must be in format partnerId:privateKeyPEM")
	}
	return newWithParts(credential[:idx], credential[idx+1:], logger)
}

func newWithParts(partnerID, privateKeyPEM string, logger *logfx.Logger) (*Client, error) {
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("eco: %w", err)
	}
	return &Client{
		BaseClient: platform.NewBaseClient(platform.PlatformECO, "https://openapi.ecosteam.cn", logger),
		partnerID:  partnerID,
		privateKey: privateKey,
	}, nil
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("eco: verifying")
	body, err := c.doRequest(ctx, "/Api/Merchant/GetTotalMoney", nil)
	if err != nil {
		c.Log.Warn("eco: verify failed", "err", err)
		return fmt.Errorf("eco verify: %w", err)
	}

	var result merchantMoneyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("eco verify: %w", err)
	}
	if result.ResultCode != "0" {
		return fmt.Errorf("eco verify: resultCode=%s", result.ResultCode)
	}
	c.Log.Info("eco: verify ok")
	return nil
}

func (c *Client) GetBalance(ctx context.Context) (*platform.Balance, error) {
	c.Log.Info("eco: fetching balance")
	body, err := c.doRequest(ctx, "/Api/Merchant/GetTotalMoney", nil)
	if err != nil {
		return nil, fmt.Errorf("eco balance: %w", err)
	}

	var result merchantMoneyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("eco balance: %w", err)
	}
	if result.ResultCode != "0" {
		return nil, fmt.Errorf("eco balance: resultCode=%s", result.ResultCode)
	}
	return &platform.Balance{
		Available: result.ResultData.Money,
		Purchase:  result.ResultData.PurchaseMoney,
		Frozen:    result.ResultData.LockMoney,
	}, nil
}

func (c *Client) GetBillHistory(_ context.Context, _ ...platform.QueryOption) ([]platform.BillRecord, error) {
	return nil, nil
}

func (c *Client) GetBuyHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("eco: fetching buy history", "since", cfg.Since)
	trades, err := c.fetchHistory(ctx, "buy", cfg, c.fetchBuyPage)
	if err != nil {
		return trades, err
	}
	c.Log.Info("eco: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) GetSellHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("eco: fetching sell history", "since", cfg.Since)
	trades, err := c.fetchHistory(ctx, "sell", cfg, c.fetchSellPage)
	if err != nil {
		return trades, err
	}
	c.Log.Info("eco: sell history done", "total", len(trades))
	return trades, nil
}

type pageFetchFn func(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string, windowStart, windowEnd time.Time) ([]platform.TradeRecord, bool, error)

func (c *Client) fetchHistory(
	ctx context.Context,
	direction string,
	cfg platform.QueryConfig,
	fetchPage pageFetchFn,
) ([]platform.TradeRecord, error) {
	var all []platform.TradeRecord
	windowEnd := time.Now()

	var sinceTime time.Time
	if cfg.Since > 0 {
		sinceTime = time.UnixMilli(cfg.Since)
	}

	consecutiveEmpty := 0
	for {
		windowStart := windowEnd.AddDate(0, 0, -apiMaxDays)
		if !sinceTime.IsZero() && windowStart.Before(sinceTime) {
			windowStart = sinceTime
		}
		if !windowEnd.After(windowStart) {
			break
		}

		remaining := cfg.Limit
		if remaining > 0 {
			remaining -= len(all)
			if remaining <= 0 {
				break
			}
		}

		trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, direction, 500*time.Millisecond, remaining,
			func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
				return fetchPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams, windowStart, windowEnd)
			},
		)
		if err != nil {
			return all, err
		}
		all = append(all, trades...)
		c.Log.Info("eco: time window done", "direction", direction,
			"window", windowStart.Format(dateFormat)+"~"+windowEnd.Format(dateFormat),
			"count", len(trades), "total", len(all))

		if !sinceTime.IsZero() && !windowStart.After(sinceTime) {
			break
		}

		if len(trades) == 0 {
			consecutiveEmpty++
			if consecutiveEmpty >= 12 {
				break
			}
		} else {
			consecutiveEmpty = 0
		}
		windowEnd = windowStart
		time.Sleep(500 * time.Millisecond)
	}

	return all, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string, windowStart, windowEnd time.Time) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"StartTime": windowStart.Format(dateFormat),
		"EndTime":   windowEnd.Format(dateFormat),
		"PageIndex": page,
		"PageSize":  100,
	}
	if tradeState == platform.TradeStateCompleted {
		body["OrderState"] = OrderStateSuccess
	}
	for k, v := range extra {
		body[k] = v
	}

	reqBody, _ := json.Marshal(body)
	c.Log.Debug("eco: buyer order list request", "body", string(reqBody))

	respBody, err := c.doRequest(ctx, "/Api/open/order/BuyerOrderList", body)
	if err != nil {
		return nil, false, err
	}

	var result buyerOrderListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, err
	}
	if result.ResultCode != "0" {
		return nil, false, fmt.Errorf("eco API error: code=%s msg=%s", result.ResultCode, result.ResultMsg)
	}

	c.Log.Debug("eco: buyer order list response", "totalRecord", result.ResultData.TotalRecord, "pageSize", result.ResultData.PageSize, "pageLen", len(result.ResultData.PageResult), "raw", string(respBody))

	trades := make([]platform.TradeRecord, 0, len(result.ResultData.PageResult))
	anyPastSince := false
	for _, o := range result.ResultData.PageResult {
		tradeAt := parseTradeAt(o.CreateOrderTime)
		if tradeAt < since {
			anyPastSince = true
			continue
		}
		trades = append(trades, toBuyTradeFromListItem(o))
	}

	c.enrichBuyPage(ctx, trades)

	hasMore := page*100 < result.ResultData.TotalRecord
	if anyPastSince {
		hasMore = false
	}

	return trades, hasMore, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string, windowStart, windowEnd time.Time) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"StartTime": windowStart.Format(dateFormat),
		"EndTime":   windowEnd.Format(dateFormat),
		"PageIndex": page,
		"PageSize":  100,
	}
	if tradeState == platform.TradeStateCompleted {
		body["OrderState"] = OrderStateSuccess
	}
	for k, v := range extra {
		body[k] = v
	}

	reqBody, _ := json.Marshal(body)
	c.Log.Debug("eco: seller order list request", "body", string(reqBody))

	respBody, err := c.doRequest(ctx, "/Api/open/order/SellerOrderList", body)
	if err != nil {
		return nil, false, err
	}

	var result sellerOrderListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, err
	}
	if result.ResultCode != "0" {
		return nil, false, fmt.Errorf("eco API error: code=%s msg=%s", result.ResultCode, result.ResultMsg)
	}

	c.Log.Debug("eco: seller order list response", "totalRecord", result.ResultData.TotalRecord, "pageLen", len(result.ResultData.PageResult), "raw", string(respBody))

	trades := make([]platform.TradeRecord, 0, len(result.ResultData.PageResult))
	anyPastSince := false
	for _, o := range result.ResultData.PageResult {
		tradeAt := parseTradeAt(o.CreateOrderTime)
		if tradeAt < since {
			anyPastSince = true
			continue
		}
		trades = append(trades, toSellTradeFromListItem(o))
	}

	c.enrichSellPage(ctx, trades)

	hasMore := page*100 < result.ResultData.TotalRecord
	if anyPastSince {
		hasMore = false
	}

	return trades, hasMore, nil
}

const detailConcurrency = 5

func (c *Client) enrichBuyPage(ctx context.Context, trades []platform.TradeRecord) {
	sem := make(chan struct{}, detailConcurrency)
	var wg sync.WaitGroup

	for i := range trades {
		orderNum := trades[i].ExternalID[len("eco-buy-"):]
		wg.Add(1)
		go func(idx int, on string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			detail, err := c.fetchBuyDetail(ctx, on)
			if err != nil {
				c.Log.Debug("eco: skip buy detail", "orderNum", on, "err", err)
				return
			}
			mergeDetail(&trades[idx], &detail)
		}(i, orderNum)
	}
	wg.Wait()
}

func (c *Client) enrichSellPage(ctx context.Context, trades []platform.TradeRecord) {
	sem := make(chan struct{}, detailConcurrency)
	var wg sync.WaitGroup

	for i := range trades {
		orderNum := trades[i].ExternalID[len("eco-sell-"):]
		wg.Add(1)
		go func(idx int, on string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			detail, err := c.fetchSellDetail(ctx, on)
			if err != nil {
				c.Log.Debug("eco: skip sell detail", "orderNum", on, "err", err)
				return
			}
			mergeDetail(&trades[idx], &detail)
		}(i, orderNum)
	}
	wg.Wait()
}

func mergeDetail(t *platform.TradeRecord, d *orderDetailModel) {
	item := toCS2Item(d.AssetPreviewModel)
	if item.ItemName == "" {
		item.ItemName, item.Exterior = platform.NormalizeItemName(d.GoodsName)
	}
	if item.MarketHashName == "" {
		item.MarketHashName = d.HashName
	}
	t.CS2Item = item
	t.Fee = yuanToFen(d.BuyerFee)
	t.UnitPrice = yuanToFen(d.TotalMoney)
	t.TotalPrice = yuanToFen(d.TotalMoney)
	if d.TradeOfferId != "" {
		t.TradeOfferID = d.TradeOfferId
	}
}

func (c *Client) fetchBuyDetail(ctx context.Context, orderNum string) (orderDetailModel, error) {
	body := map[string]any{"OrderNum": orderNum}
	respBody, err := c.doRequest(ctx, "/Api/open/order/OrderDetailsInfo", body)
	if err != nil {
		return orderDetailModel{}, err
	}
	return parseOrderDetailResponse(respBody)
}

func (c *Client) fetchSellDetail(ctx context.Context, orderNum string) (orderDetailModel, error) {
	body := map[string]any{"OrderNum": orderNum}
	respBody, err := c.doRequest(ctx, "/Api/open/order/SellerOrderDetail", body)
	if err != nil {
		return orderDetailModel{}, err
	}
	return parseOrderDetailResponse(respBody)
}

func parseOrderDetailResponse(body []byte) (orderDetailModel, error) {
	var result orderDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return orderDetailModel{}, err
	}
	if result.ResultCode != "0" {
		return orderDetailModel{}, fmt.Errorf("resultCode=%s msg=%s", result.ResultCode, result.ResultMsg)
	}
	return result.ResultData, nil
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	return h
}

func (c *Client) doRequest(ctx context.Context, path string, body map[string]any) ([]byte, error) {
	if body == nil {
		body = map[string]any{}
	}
	body["PartnerId"] = c.partnerID
	body["Timestamp"] = time.Now().Unix()

	sign, err := generateRSASignature(c.privateKey, body)
	if err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}
	body["Sign"] = sign

	bodyBytes, _ := json.Marshal(body)
	_, respBody, err := c.DoRequest(ctx, "POST", path, nil, bodyBytes, c.headers())
	return respBody, err
}
