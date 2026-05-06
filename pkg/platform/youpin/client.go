package youpin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils"
)

type Client struct {
	platform.BaseClient
	token       string
	deviceToken string
	deviceID    string
	ukValue     string
	ukTime      time.Time
	userID      int64
}

func New(token string) *Client {
	dev := randomString(24)
	c := &Client{
		BaseClient:  platform.NewBaseClient(utils.PlatformYoupin, "https://api.youpin898.com"),
		token:       token,
		deviceToken: dev,
		deviceID:    dev,
	}
	c.init()
	return c
}

// init registers the device with YouPin.
func (c *Client) init() {
	var result struct {
		Code int `json:"Code"`
	}
	if err := c.doGET(context.Background(), "/api/common/ClientInfo/AndroidInfo", map[string]string{
		"DeviceToken": c.deviceToken,
		"Sessionid":   c.deviceToken,
	}, &result); err != nil {
		c.Log.Warn("youpin: init device info failed", "err", err)
		return
	}
	c.Log.Debug("youpin: init device info ok", "code", result.Code)
}

func (c *Client) Verify(ctx context.Context) error {
	var result struct {
		Code int `json:"Code"`
		Data struct {
			UserID int64 `json:"UserId"`
		} `json:"Data"`
	}
	if err := c.doGET(ctx, "/api/user/Account/getUserInfo", nil, &result); err != nil {
		c.Log.Warn("youpin verify failed", "err", err)
		return fmt.Errorf("youpin verify: %w", err)
	}
	if result.Code != 0 {
		c.Log.Warn("youpin verify: invalid token", "code", result.Code)
		return fmt.Errorf("youpin verify: token invalid (code=%d)", result.Code)
	}
	c.userID = result.Data.UserID
	c.Log.Debug("youpin verify ok")
	return nil
}

func (c *Client) FetchBuyHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Debug("youpin: fetching buy history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "buy", 1*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchBuyPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("youpin: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"keys":        "",
		"orderStatus": 340,
		"pageIndex":   page,
		"pageSize":    20,
		"presenterId": 0,
		"sceneType":   0,
		"Sessionid":   c.deviceID,
	}

	var result youpinBuyPageResponse
	if err := c.doPOST(ctx, "/api/youpin/bff/trade/sale/v1/buy/list", body, &result); err != nil {
		c.Log.Warn("youpin buy page failed", "page", page, "err", err)
		return nil, false, err
	}
	if result.Code != 0 {
		c.Log.Warn("youpin buy page: API error", "code", result.Code)
		return nil, false, fmt.Errorf("youpin API error: code=%d", result.Code)
	}

	c.Log.Debug("youpin buy page", "page", page, "orders", len(result.Data.OrderList), "totalCount", result.Data.TotalCount)

	trades := make([]platform.TradeRecord, 0, len(result.Data.OrderList))
	for _, o := range result.Data.OrderList {
		if o.OrderStatusName != "已完成" {
			continue
		}
		if o.FinishOrderTime < since {
			continue
		}
		// Batch orders (commodityNum > 3) need a separate detail API call.
		if o.CommodityNum > 3 {
			batchTrades, err := c.fetchBuyBatch(ctx, o.ID, o.BuyerUserID)
			if err != nil {
				c.Log.Warn("youpin buy batch failed", "orderID", o.ID, "err", err)
				continue
			}
			trades = append(trades, batchTrades...)
			continue
		}
		for _, p := range o.ProductList {
			trades = append(trades, c.buildBuyTrade(o, p))
		}
	}

	hasMore := page*20 < result.Data.TotalCount
	return trades, hasMore, nil
}

func (c *Client) buildBuyTrade(o youpinBuyOrder, p youpinBuyProduct) platform.TradeRecord {
	// AssetID: prefer assertId, fallback to commodityId.
	assetID := fmt.Sprintf("%d", p.AssertID)
	if p.AssertID == 0 {
		assetID = fmt.Sprintf("%d", p.CommodityID)
	}

	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("youpin-buy-%d-%d", o.OrderID, p.AssertID),
		AssetID:    assetID,
		ItemName:   p.CommodityName,
		TradeType:  "buy",
		Quantity:   1,
		UnitPrice:  p.Price,
		TotalPrice: p.Price,
		Fee:        0,
		TradeAt:    o.FinishOrderTime,
	}
}

func (c *Client) fetchBuyBatch(ctx context.Context, orderID int64, buyerUserID int64) ([]platform.TradeRecord, error) {
	body := map[string]any{
		"orderNo":   fmt.Sprintf("%d", orderID),
		"userId":    buyerUserID,
		"Sessionid": c.deviceID,
	}

	type commodityVO struct {
		ID     int64   `json:"id"`
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
		Abrade string  `json:"abrade"`
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			OrderCanceledTime   int64 `json:"orderCanceledTime"`
			UserCommodityVOList []struct {
				CommodityVOList []commodityVO `json:"commodityVOList"`
			} `json:"userCommodityVOList"`
		} `json:"data"`
	}

	if err := c.doPOST(ctx, "/api/youpin/bff/trade/v1/order/query/detail", body, &result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("batch order API error: code=%d", result.Code)
	}

	var trades []platform.TradeRecord
	if len(result.Data.UserCommodityVOList) > 0 {
		for _, commodity := range result.Data.UserCommodityVOList[0].CommodityVOList {
			// Batch detail API returns price in yuan; TradeRecord stores fen.
			priceFen := int64(commodity.Price * 100)
			trades = append(trades, platform.TradeRecord{
				ExternalID: fmt.Sprintf("youpin-buy-%d-%d", orderID, commodity.ID),
				AssetID:    fmt.Sprintf("%d", commodity.ID),
				ItemName:   commodity.Name,
				TradeType:  "buy",
				Quantity:   1,
				UnitPrice:  priceFen,
				TotalPrice: priceFen,
				Fee:        0,
				TradeAt:    result.Data.OrderCanceledTime,
			})
		}
	}
	return trades, nil
}

func (c *Client) FetchSellHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Debug("youpin: fetching sell history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "sell", 1*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("youpin: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"keys":        "",
		"orderStatus": "340",
		"pageIndex":   page,
		"pageSize":    20,
	}

	var result youpinSellPageResponse
	if err := c.doPOST(ctx, "/api/youpin/bff/trade/sale/v1/sell/list", body, &result); err != nil {
		c.Log.Warn("youpin sell page failed", "page", page, "err", err)
		return nil, false, err
	}
	if result.Code != 0 {
		c.Log.Warn("youpin sell page: API error", "code", result.Code)
		return nil, false, fmt.Errorf("youpin API error: code=%d", result.Code)
	}

	c.Log.Debug("youpin sell page", "page", page, "orders", len(result.Data.OrderList), "totalCount", result.Data.TotalCount)

	trades := make([]platform.TradeRecord, 0, len(result.Data.OrderList))
	for _, o := range result.Data.OrderList {
		if o.FinishOrderTime < since {
			continue
		}
		qty := int64(o.CommodityNum)
		if qty == 0 {
			qty = 1
		}
		trades = append(trades, platform.TradeRecord{
			ExternalID: fmt.Sprintf("youpin-sell-%s", o.OrderNo),
			AssetID:    fmt.Sprintf("%d", o.ProductDetail.AssertID),
			ItemName:   o.ProductDetail.CommodityName,
			TradeType:  "sell",
			Quantity:   qty,
			UnitPrice:  o.PaymentAmount / qty,
			TotalPrice: o.PaymentAmount,
			Fee:        0,
			TradeAt:    o.FinishOrderTime,
		})
	}

	hasMore := page*20 < result.Data.TotalCount
	return trades, hasMore, nil
}

func (c *Client) FetchBalance(ctx context.Context) (*platform.Balance, error) {
	return &platform.Balance{}, nil
}

// setHeaders sets all YouPin API headers. If pcPlatform is true, uses PC platform
// instead of Android. If ukVerify is true, ensures a fresh UK (30s TTL) is present.
func (c *Client) setHeaders(req *http.Request, pcPlatform, ukVerify bool) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "okhttp/3.14.9")
	req.Header.Set("App-Version", "5.28.3")
	req.Header.Set("AppType", "4")
	req.Header.Set("deviceType", "1")
	req.Header.Set("package-type", "uuyp")
	req.Header.Set("DeviceToken", c.deviceToken)
	req.Header.Set("DeviceId", c.deviceID)
	req.Header.Set("Gameid", "730")

	if pcPlatform {
		req.Header.Set("platform", "pc")
	} else {
		req.Header.Set("platform", "android")
	}

	if ukVerify {
		c.refreshUKIfNeeded()
		req.Header.Set("uk", c.ukValue)
	}

	deviceInfo := fmt.Sprintf(
		`{"deviceId":"%s","deviceType":"%s","hasSteamApp":1,"requestTag":"%s","systemName":"Android","systemVersion":"15"}`,
		c.deviceID, c.deviceID, strings.ToUpper(randomString(32)),
	)
	req.Header.Set("Device-Info", deviceInfo)
}

// refreshUKIfNeeded fetches a fresh UK from /api/deviceW2 if the cached one is
// older than 30 seconds.
func (c *Client) refreshUKIfNeeded() {
	if c.ukValue != "" && time.Since(c.ukTime) < 30*time.Second {
		return
	}

	uk, err := c.fetchUK()
	if err != nil {
		c.Log.Warn("youpin: fetch UK failed, using random", "err", err)
		c.ukValue = randomString(65)
		c.ukTime = time.Time{}
		return
	}
	c.ukValue = uk
	c.ukTime = time.Now()
	c.Log.Debug("youpin: UK refreshed", "expires_in", "30s")
}

func (c *Client) doGET(ctx context.Context, path string, query map[string]string, result any) error {
	url := c.BaseURL + path
	if len(query) > 0 {
		params := make([]string, 0, len(query))
		for k, v := range query {
			params = append(params, k+"="+v)
		}
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req, false, false)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	c.Log.Debug("youpin API response", "method", "GET", "path", path, "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("youpin API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	if err := c.checkAPIError(respBody); err != nil {
		return err
	}
	return json.Unmarshal(respBody, result)
}

func (c *Client) doPOST(ctx context.Context, path string, body map[string]any, result any) error {
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.setHeaders(req, false, false)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	c.Log.Debug("youpin API response", "method", "POST", "path", path, "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("youpin API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	if err := c.checkAPIError(respBody); err != nil {
		return err
	}
	return json.Unmarshal(respBody, result)
}

// checkAPIError checks for known YouPin error codes in the response.
// Handles both "code" and "Code" since the API is inconsistent.
func (c *Client) checkAPIError(body []byte) error {
	var resp struct {
		Code  int `json:"Code"`
		CodeL int `json:"code"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil // can't parse, let the caller handle
	}

	code := resp.Code
	if code == 0 {
		code = resp.CodeL
	}

	switch code {
	case 84101:
		return fmt.Errorf("youpin: 登录状态失效，请重新登录 (session expired)")
	case 84104:
		return fmt.Errorf("youpin: 风控限制，暂时无法访问 (rate limited)")
	case 9004001:
		// Empty list — not an error, just no data.
		return nil
	}
	if code != 0 {
		c.Log.Warn("youpin: API returned non-zero code", "code", code)
	}
	return nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
