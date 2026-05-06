package igxe

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils"
)

type Client struct {
	platform.BaseClient
	partnerID  string
	privateKey *rsa.PrivateKey
}

// New accepts credential in format "partnerId:privateKeyPEM".
func New(credential string) (*Client, error) {
	idx := strings.Index(credential, ":")
	if idx < 0 {
		return nil, fmt.Errorf("igxe: credential must be in format partnerId:privateKeyPEM")
	}
	return newWithParts(credential[:idx], credential[idx+1:])
}

func newWithParts(partnerID, privateKeyPEM string) (*Client, error) {
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("igxe: %w", err)
	}
	return &Client{
		BaseClient: platform.NewBaseClient(utils.PlatformIGXE, "https://openapi.ecosteam.cn"),
		partnerID:  partnerID,
		privateKey: privateKey,
	}, nil
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("igxe: verifying")
	_, err := c.getTotalMoney(ctx)
	if err != nil {
		c.Log.Warn("igxe: verify failed", "err", err)
		return fmt.Errorf("igxe verify: %w", err)
	}
	c.Log.Info("igxe: verify ok")
	return nil
}

func (c *Client) FetchBalance(ctx context.Context) (*platform.Balance, error) {
	c.Log.Info("igxe: fetching balance")
	data, err := c.getTotalMoney(ctx)
	if err != nil {
		return nil, fmt.Errorf("igxe balance: %w", err)
	}
	return &platform.Balance{
		Available: int64(data.ResultData.Money * 100),
		Purchase:  0,
	}, nil
}

func (c *Client) FetchBuyHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	return []platform.TradeRecord{}, nil
}

func (c *Client) FetchSellHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Info("igxe: fetching sell history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "sell", 1*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("igxe: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) getTotalMoney(ctx context.Context) (*totalMoneyResponse, error) {
	body := map[string]any{}
	var result totalMoneyResponse
	if err := c.doPOST(ctx, "/Api/Merchant/GetTotalMoney", body, &result); err != nil {
		return nil, err
	}
	if result.ResultCode != "0" {
		return nil, fmt.Errorf("API error: resultCode=%s", result.ResultCode)
	}
	return &result, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	startTime := time.UnixMilli(since).Format("2006-01-02")
	endTime := time.Now().Format("2006-01-02")

	body := map[string]any{
		"StartTime": startTime,
		"EndTime":   endTime,
		"PageIndex": page,
		"PageSize":  100,
	}

	var result sellerOrderListResponse
	if err := c.doPOST(ctx, "/Api/open/order/SellerOrderList", body, &result); err != nil {
		return nil, false, err
	}
	if result.ResultCode != "0" {
		return nil, false, fmt.Errorf("igxe API error: resultCode=%s msg=%s", result.ResultCode, result.ResultMsg)
	}

	trades := make([]platform.TradeRecord, 0, len(result.ResultData.PageResult))
	for _, item := range result.ResultData.PageResult {
		tradeAt := c.parseCreateDate(item.CreateDate)
		if tradeAt < since {
			continue
		}
		price := int64(item.Price * 100)
		trades = append(trades, platform.TradeRecord{
			ExternalID: fmt.Sprintf("igxe-sell-%s", item.OrderNum),
			AssetID:    item.AssetID,
			ItemName:   item.GoodsName,
			TradeType:  "sell",
			Quantity:   1,
			UnitPrice:  price,
			TotalPrice: price,
			Fee:        0,
			TradeAt:    tradeAt,
		})
	}

	hasMore := len(result.ResultData.PageResult) >= 100
	return trades, hasMore, nil
}

func (c *Client) parseCreateDate(s string) int64 {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t.UnixMilli()
		}
	}
	c.Log.Warn("igxe: failed to parse create date", "value", s)
	return 0
}

func (c *Client) doPOST(ctx context.Context, path string, body map[string]any, result any) error {
	body["PartnerId"] = c.partnerID
	body["Timestamp"] = time.Now().Unix()

	sign, err := generateRSASignature(c.privateKey, body)
	if err != nil {
		return fmt.Errorf("sign error: %w", err)
	}
	body["Sign"] = sign

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CS2Ledger")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.Log.Debug("igxe API response", "path", path, "status", resp.StatusCode, "body", string(respBody))

	return json.Unmarshal(respBody, result)
}
