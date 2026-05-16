package igxe

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

const dateFormat = "2006-01-02 15:04"

var tradeAtFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02",
}

type Client struct {
	platform.BaseClient
	partnerID  string
	privateKey *rsa.PrivateKey
}

// New accepts credential in format "partnerId:privateKeyPEM".
func New(credential string, logger *logfx.Logger) (*Client, error) {
	idx := strings.Index(credential, ":")
	if idx < 0 {
		return nil, fmt.Errorf("igxe: credential must be in format partnerId:privateKeyPEM")
	}
	return newWithParts(credential[:idx], credential[idx+1:], logger)
}

func newWithParts(partnerID, privateKeyPEM string, logger *logfx.Logger) (*Client, error) {
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("igxe: %w", err)
	}
	return &Client{
		BaseClient: platform.NewBaseClient(platform.PlatformIGXE, "https://openapi.ecosteam.cn", logger),
		partnerID:  partnerID,
		privateKey: privateKey,
	}, nil
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("igxe: verifying")
	_, body, err := c.doRequest(ctx, "POST", "/Api/Merchant/GetTotalMoney", nil, nil)
	if err != nil {
		c.Log.Warn("igxe: verify failed", "err", err)
		return fmt.Errorf("igxe verify: %w", err)
	}

	var result totalMoneyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("igxe verify: %w", err)
	}
	if result.ResultCode != "0" {
		return fmt.Errorf("igxe verify: resultCode=%s", result.ResultCode)
	}
	c.Log.Info("igxe: verify ok")
	return nil
}

func (c *Client) GetBalance(ctx context.Context) (*platform.Balance, error) {
	c.Log.Info("igxe: fetching balance")
	_, body, err := c.doRequest(ctx, "POST", "/Api/Merchant/GetTotalMoney", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("igxe balance: %w", err)
	}

	var result totalMoneyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("igxe balance: %w", err)
	}
	if result.ResultCode != "0" {
		return nil, fmt.Errorf("igxe balance: resultCode=%s", result.ResultCode)
	}
	return &platform.Balance{
		Available: result.ResultData.Money,
		Purchase:  0,
	}, nil
}

func (c *Client) GetBuyHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	return []platform.TradeRecord{}, nil
}

func (c *Client) GetSellHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("igxe: fetching sell history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "sell", 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("igxe: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string) ([]platform.TradeRecord, bool, error) {
	_ = tradeState
	body := map[string]any{
		"StartTime": time.Now().Add(-30 * 24 * time.Hour).Format(dateFormat),
		"EndTime":   time.Now().Format(dateFormat),
		"PageIndex": page,
		"PageSize":  100,
	}
	for k, v := range extra {
		body[k] = v
	}
	_, respBody, err := c.doRequest(ctx, "POST", "/Api/open/order/SellerOrderList", nil, body)
	if err != nil {
		return nil, false, err
	}

	var result sellerOrderListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, err
	}
	if result.ResultCode != "0" {
		return nil, false, fmt.Errorf("igxe API error: resultCode=%s msg=%s", result.ResultCode, result.ResultMsg)
	}

	trades := make([]platform.TradeRecord, 0, len(result.ResultData.PageResult))
	anyAfterSince := false
	for _, item := range result.ResultData.PageResult {
		tradeAt := c.parseCreateDate(item.CreateDate)
		if tradeAt >= since {
			anyAfterSince = true
		}
		if tradeAt < since {
			continue
		}
		trades = append(trades, toSellTrade(item, tradeAt))
	}

	if len(result.ResultData.PageResult) > 0 && !anyAfterSince {
		return trades, false, nil
	}
	hasMore := len(result.ResultData.PageResult) >= 100
	return trades, hasMore, nil
}

func (c *Client) parseCreateDate(s string) int64 {
	for _, f := range tradeAtFormats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t.UnixMilli()
		}
	}
	c.Log.Warn("igxe: failed to parse create date", "value", s)
	return 0
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("User-Agent", "CS2Ledger")
	return h
}

//nolint:unparam // unified signature — status code not always checked by callers
func (c *Client) doRequest(ctx context.Context, method, path string, query map[string]string, body map[string]any) (int, []byte, error) {
	if body == nil {
		body = map[string]any{}
	}
	body["PartnerId"] = c.partnerID
	body["Timestamp"] = time.Now().Unix()

	sign, err := generateRSASignature(c.privateKey, body)
	if err != nil {
		return 0, nil, fmt.Errorf("sign error: %w", err)
	}
	body["Sign"] = sign

	bodyBytes, _ := json.Marshal(body)
	return c.DoRequest(ctx, method, path, query, bodyBytes, c.headers())
}
