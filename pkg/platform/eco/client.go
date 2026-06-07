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

// parseECO unmarshals an ecoResponse envelope and checks ResultCode.
func parseECO[T any](body []byte) (T, error) {
	var result ecoResponse[T]
	if err := json.Unmarshal(body, &result); err != nil {
		var zero T
		return zero, err
	}
	if result.ResultCode != "0" {
		var zero T
		return zero, fmt.Errorf("eco API error: code=%s msg=%s", result.ResultCode, result.ResultMsg)
	}
	return result.ResultData, nil
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("eco: verifying")
	body, err := c.doRequest(ctx, "/Api/Merchant/GetTotalMoney", nil)
	if err != nil {
		c.Log.Warn("eco: verify failed", "err", err)
		return fmt.Errorf("eco verify: %w", err)
	}

	if _, err := parseECO[merchantMoneyModel](body); err != nil {
		return fmt.Errorf("eco verify: %w", err)
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

	data, err := parseECO[merchantMoneyModel](body)
	if err != nil {
		return nil, fmt.Errorf("eco balance: %w", err)
	}
	return &platform.Balance{
		Available: data.Money,
		Purchase:  data.PurchaseMoney,
		Frozen:    data.LockMoney,
	}, nil
}

func (c *Client) GetBillHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.BillRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("eco: fetching bill history", "since", cfg.Since)

	bills, err := platform.FetchByTimeWindows(ctx, c.Log, c.Name, "bill", cfg, apiMaxDays,
		func(ctx context.Context, page int, windowStart, windowEnd time.Time) ([]platform.BillRecord, bool, error) {
			return c.fetchBillPage(ctx, page, cfg.Since, windowStart, windowEnd)
		},
	)
	if err != nil {
		return bills, err
	}
	c.Log.Info("eco: bill history done", "total", len(bills))
	return bills, nil
}

func (c *Client) fetchBillPage(ctx context.Context, page int, since int64, windowStart, windowEnd time.Time) ([]platform.BillRecord, bool, error) {
	body := map[string]any{
		"PageIndex": page,
		"PageSize":  100,
		"StartTime": windowStart.Format(dateFormat),
		"EndTime":   windowEnd.Format(dateFormat),
	}
	respBody, err := c.doRequest(ctx, "/Api/Merchant/GetFundFlow", body)
	if err != nil {
		return nil, false, err
	}

	c.Log.Debug("eco: fund flow raw response", "body", string(respBody))

	pages, err := parseECO[fundFlowPagesModel](respBody)
	if err != nil {
		return nil, false, err
	}

	c.Log.Debug("eco: fund flow response parsed", "page", page, "totalRecord", pages.TotalRecord, "pageResultLen", len(pages.PageResult), "pageSize", pages.PageSize)

	records := make([]platform.BillRecord, 0, len(pages.PageResult))
	finished := false
	for _, item := range pages.PageResult {
		rec, err := toBillRecord(item)
		if err != nil {
			c.Log.Warn("eco: bill item parse failed, skipping", "err", err)
			continue
		}
		if since > 0 && rec.AddTime < since {
			finished = true
			continue
		}
		records = append(records, rec)
	}

	c.Log.Debug("eco: fund flow processed", "page", page, "recordsAfterFilter", len(records), "finished", finished, "since", since)

	hasMore := !finished && len(pages.PageResult) == pages.PageSize
	if pages.TotalRecord > 0 {
		hasMore = page*pages.PageSize < pages.TotalRecord && !finished
	}
	return records, hasMore, nil
}

func (c *Client) GetBuyHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("eco: fetching buy history", "since", cfg.Since)
	trades, err := platform.FetchByTimeWindows(ctx, c.Log, c.Name, "buy", cfg, apiMaxDays,
		func(ctx context.Context, page int, windowStart, windowEnd time.Time) ([]platform.TradeRecord, bool, error) {
			return c.fetchBuyPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams, windowStart, windowEnd)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("eco: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) GetSellHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("eco: fetching sell history", "since", cfg.Since)
	trades, err := platform.FetchByTimeWindows(ctx, c.Log, c.Name, "sell", cfg, apiMaxDays,
		func(ctx context.Context, page int, windowStart, windowEnd time.Time) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams, windowStart, windowEnd)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("eco: sell history done", "total", len(trades))
	return trades, nil
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

	pages, err := parseECO[buyerOrderPagesModel](respBody)
	if err != nil {
		return nil, false, err
	}

	c.Log.Debug("eco: buyer order list response", "totalRecord", pages.TotalRecord, "pageSize", pages.PageSize, "pageLen", len(pages.PageResult), "raw", string(respBody))

	trades := make([]platform.TradeRecord, 0, len(pages.PageResult))
	anyPastSince := false
	for _, o := range pages.PageResult {
		tradeAt := parseTradeAt(o.CreateOrderTime)
		if tradeAt < since {
			anyPastSince = true
			continue
		}
		trades = append(trades, toBuyTradeFromListItem(o))
	}

	c.enrichBuyPage(ctx, trades)

	hasMore := page*100 < pages.TotalRecord
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

	pages, err := parseECO[sellerOrderPagesModel](respBody)
	if err != nil {
		return nil, false, err
	}

	c.Log.Debug("eco: seller order list response", "totalRecord", pages.TotalRecord, "pageLen", len(pages.PageResult), "raw", string(respBody))

	trades := make([]platform.TradeRecord, 0, len(pages.PageResult))
	anyPastSince := false
	for _, o := range pages.PageResult {
		tradeAt := parseTradeAt(o.CreateOrderTime)
		if tradeAt < since {
			anyPastSince = true
			continue
		}
		trades = append(trades, toSellTradeFromListItem(o))
	}

	c.enrichSellPage(ctx, trades)

	hasMore := page*100 < pages.TotalRecord
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
	return parseECO[orderDetailModel](respBody)
}

func (c *Client) fetchSellDetail(ctx context.Context, orderNum string) (orderDetailModel, error) {
	body := map[string]any{"OrderNum": orderNum}
	respBody, err := c.doRequest(ctx, "/Api/open/order/SellerOrderDetail", body)
	if err != nil {
		return orderDetailModel{}, err
	}
	return parseECO[orderDetailModel](respBody)
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
