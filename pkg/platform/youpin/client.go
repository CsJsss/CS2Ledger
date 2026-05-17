package youpin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/httpclient"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

const (
	DefaultPageSize int = 20
)

type Client struct {
	platform.BaseClient
	token       string
	deviceToken string
	deviceID    string
	userID      int64
	initOnce    sync.Once
	uk          string
	ukTime      time.Time
	ukMu        sync.Mutex
}

func New(token string, logger *logfx.Logger) *Client {
	dev := randomString(24)
	return &Client{
		BaseClient: platform.NewBaseClient(
			platform.PlatformYoupin, "https://api.youpin898.com", logger,
			httpclient.WithRateLimit(3, 2),
		),
		token:       token,
		deviceToken: dev,
		deviceID:    dev,
	}
}

func (c *Client) ensureUK(ctx context.Context) {
	c.ukMu.Lock()
	uk := c.uk
	ukTime := c.ukTime
	c.ukMu.Unlock()

	if uk != "" && time.Since(ukTime) < 30*time.Second {
		return
	}

	fetched, err := c.fetchUK(ctx)
	if err != nil {
		c.Log.Warn("youpin: fetch UK failed, using random fallback", "err", err)
		fetched = randomString(65)
	}

	c.ukMu.Lock()
	c.uk = fetched
	c.ukTime = time.Now()
	c.ukMu.Unlock()
	c.Log.Debug("youpin: UK refreshed")
}

func (c *Client) fetchUK(ctx context.Context) (string, error) {
	crypt, err := newUUApiCrypt(randomString(16))
	if err != nil {
		return "", fmt.Errorf("fetch UK: %w", err)
	}

	data := `{"iud":"` + randomUUID() + `"}`
	encryptedData, err := crypt.uuEncrypt(data)
	if err != nil {
		return "", fmt.Errorf("fetch UK: encrypt: %w", err)
	}
	encryptedAesKey, err := crypt.getEncryptedAesKey()
	if err != nil {
		return "", fmt.Errorf("fetch UK: encrypted aes key: %w", err)
	}

	payload := map[string]string{
		"encryptedData":   encryptedData,
		"encryptedAesKey": encryptedAesKey,
	}
	body, _ := json.Marshal(payload)

	status, respBody, err := c.DoRequest(ctx, "POST", "/api/deviceW2", nil, body, c.headers())
	if err != nil {
		return "", fmt.Errorf("fetch UK: %w", err)
	}
	if status != 200 {
		return "", fmt.Errorf("fetch UK: HTTP %d: %s", status, string(respBody))
	}

	plain, err := crypt.uuDecrypt(string(respBody))
	if err != nil {
		return "", fmt.Errorf("fetch UK: decrypt response: %w", err)
	}

	var result struct {
		U string `json:"u"`
	}
	if err := json.Unmarshal([]byte(plain), &result); err != nil {
		return "", fmt.Errorf("fetch UK: parse response: %w", err)
	}
	if result.U == "" {
		return "", fmt.Errorf("fetch UK: empty uk in response")
	}
	return result.U, nil
}

// registerDevice registers the device token with YouPin.
func (c *Client) registerDevice() {
	c.initOnce.Do(func() {
		var result struct {
			Code int `json:"Code"`
		}
		if err := c.call(context.Background(), "GET", "/api/common/ClientInfo/AndroidInfo", map[string]string{
			"DeviceToken": c.deviceToken,
			"Sessionid":   c.deviceToken,
		}, nil, &result); err != nil {
			c.Log.Warn("youpin: init device info failed", "err", err)
		}
	})
}

func (c *Client) Verify(ctx context.Context) error {
	c.registerDevice()
	_, body, err := c.doRequest(ctx, "GET", "/api/user/Account/getUserInfo", nil, nil)
	if err != nil {
		c.Log.Warn("youpin verify failed", "err", err)
		return fmt.Errorf("youpin verify: %w", err)
	}

	var result struct {
		Code int `json:"Code"`
		Data struct {
			UserID int64 `json:"UserId"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
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

func (c *Client) GetBuyHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	c.registerDevice()
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Debug("youpin: fetching buy history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, model.DirectionBuy, 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchBuyPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("youpin: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"keys":        "",
		"orderStatus": 340,
		"pageIndex":   page,
		"pageSize":    DefaultPageSize,
		"presenterId": 0,
		"sceneType":   0,
		"Sessionid":   c.deviceID,
	}
	for k, v := range extra {
		body[k] = v
	}
	_, respBody, err := c.doRequest(ctx, "POST", "/api/youpin/bff/trade/sale/v1/buy/list", nil, body)
	if err != nil {
		return nil, false, err
	}

	var result youpinBuyPageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.Log.Warn("youpin buy page failed", "page", page, "err", err)
		return nil, false, err
	}
	if result.Code != 0 {
		c.Log.Warn("youpin buy page: API error", "code", result.Code)
		return nil, false, fmt.Errorf("youpin API error: code=%d", result.Code)
	}

	c.Log.Debug("youpin buy page", "page", page, "orders", len(result.Data.OrderList), "totalCount", result.Data.TotalCount)

	trades := make([]platform.TradeRecord, 0, len(result.Data.OrderList))
	finished := false
	for _, o := range result.Data.OrderList {
		if tradeState == platform.TradeStateCompleted && o.OrderStatusName != "已完成" {
			continue
		}
		if o.FinishOrderTime < since {
			finished = true
			continue
		}
		// Batch orders (commodityNum > 3) need a separate detail API call.
		if o.CommodityNum > 3 {
			batchTrades, err := c.fetchBuyBatch(ctx, o.ID, o.BuyerUserID, o.FinishOrderTime)
			if err != nil {
				return nil, false, fmt.Errorf("youpin buy batch %s: %w", o.ID, err)
			}
			trades = append(trades, batchTrades...)
			continue
		}
		for _, p := range o.ProductList {
			trades = append(trades, toBuyTrade(o, p))
		}
	}

	if len(result.Data.OrderList) == 0 || finished {
		return trades, false, nil
	}
	// API may return null total; fall back to page-full heuristic.
	if result.Data.TotalCount > 0 {
		return trades, page*DefaultPageSize < result.Data.TotalCount, nil
	}
	return trades, len(result.Data.OrderList) == DefaultPageSize, nil
}

func (c *Client) fetchBuyBatch(ctx context.Context, orderNo string, buyerUserID int64, finishTime int64) ([]platform.TradeRecord, error) {
	body := map[string]any{
		"orderNo":   orderNo,
		"userId":    buyerUserID,
		"Sessionid": c.deviceID,
	}

	type commodityVO struct {
		ID                int64       `json:"id"`
		Name              string      `json:"name"`
		CommodityHashName string      `json:"commodityHashName"` // market hash name
		Price             json.Number `json:"price"`
		CommodityAmount   json.Number `json:"commodityAmount"`
		Abrade            string      `json:"abrade"`
		CommodityAbrade   string      `json:"commodityAbrade"`
		ExteriorName      string      `json:"exteriorName"`
		RarityName        string      `json:"rarityName"`
		ItemSetName       string      `json:"itemSetName"`
		TypeName          string      `json:"typeName"`
		PaintIndex        int         `json:"paintIndex"`
		PaintSeed         int         `json:"paintSeed"`
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			UserCommodityVOList []struct {
				CommodityVOList []commodityVO `json:"commodityVOList"`
			} `json:"userCommodityVOList"`
		} `json:"data"`
	}

	if err := c.call(ctx, "POST", "/api/youpin/bff/trade/v1/order/query/detail", nil, body, &result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("batch order API error: code=%d", result.Code)
	}

	var trades []platform.TradeRecord
	if len(result.Data.UserCommodityVOList) > 0 {
		for _, commodity := range result.Data.UserCommodityVOList[0].CommodityVOList {
			priceVal, _ := commodity.Price.Float64()
			priceFen := int64(priceVal * 100)
			totalPriceFen := priceFen
			if ca, _ := commodity.CommodityAmount.Float64(); ca > 0 {
				totalPriceFen = int64(ca * 100)
			}
			wear := commodity.CommodityAbrade
			if wear == "" {
				wear = commodity.Abrade
			}
			name, ext := platform.NormalizeItemName(commodity.Name)
			exterior := commodity.ExteriorName
			if exterior == "" {
				exterior = ext
			}
			trades = append(trades, platform.TradeRecord{
				ExternalID: fmt.Sprintf("youpin-buy-%s-%d", orderNo, commodity.ID),
				CS2Item: model.CS2Item{
					AssetID: fmt.Sprintf("%d", commodity.ID), ItemName: name, MarketHashName: commodity.CommodityHashName,
					Exterior: exterior, PaintWear: parseWear(wear),
					Rarity: commodity.RarityName, WeaponType: commodity.TypeName, Itemset: commodity.ItemSetName,
					PaintSeed: commodity.PaintSeed, PaintIndex: commodity.PaintIndex,
				},
				TradeType:  model.DirectionBuy,
				Quantity:   1,
				UnitPrice:  priceFen,
				TotalPrice: totalPriceFen,
				Fee:        0,
				TradeAt:    finishTime,
			})
		}
	}
	return trades, nil
}

func (c *Client) GetSellHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	c.registerDevice()
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Debug("youpin: fetching sell history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, model.DirectionSell, 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("youpin: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string) ([]platform.TradeRecord, bool, error) {
	body := map[string]any{
		"keys":        "",
		"orderStatus": 340,
		"pageIndex":   page,
		"pageSize":    DefaultPageSize,
		"Sessionid":   c.deviceID,
	}
	for k, v := range extra {
		body[k] = v
	}
	_, respBody, err := c.doRequest(ctx, "POST", "/api/youpin/bff/trade/sale/v1/sell/list", nil, body)
	if err != nil {
		return nil, false, err
	}

	var result youpinSellPageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.Log.Warn("youpin sell page failed", "page", page, "err", err)
		return nil, false, err
	}
	if result.Code != 0 {
		c.Log.Warn("youpin sell page: API error", "code", result.Code)
		return nil, false, fmt.Errorf("youpin API error: code=%d", result.Code)
	}

	c.Log.Debug("youpin sell page", "page", page, "orders", len(result.Data.OrderList), "totalCount", result.Data.TotalCount)

	trades := make([]platform.TradeRecord, 0, len(result.Data.OrderList))
	finished := false
	for _, o := range result.Data.OrderList {
		if tradeState == platform.TradeStateCompleted && o.OrderStatusName != "已完成" {
			continue
		}
		if o.FinishOrderTime < since {
			finished = true
			continue
		}
		trades = append(trades, toSellTrade(o))
	}

	if len(result.Data.OrderList) == 0 || finished {
		return trades, false, nil
	}
	// API may return null total
	return trades, len(result.Data.OrderList) == DefaultPageSize, nil
}

func (c *Client) GetBalance(ctx context.Context) (*platform.Balance, error) {
	c.registerDevice()
	c.ensureUK(ctx)
	c.Log.Info("youpin: fetching balance")
	_, body, err := c.doRequest(ctx, "POST", "/api/youpin/bff/payment/v1/user/account/info", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("youpin balance: %w", err)
	}

	var result youpinBalanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("youpin balance: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("youpin balance: API error code=%d", result.Code)
	}

	var data youpinBalanceData
	// The API returns data as a JSON-stringified object
	if err := json.Unmarshal(result.Data, &data); err != nil {
		var s string
		if err2 := json.Unmarshal(result.Data, &s); err2 != nil {
			return nil, fmt.Errorf("youpin balance: parse data: %w", err)
		}
		if err := json.Unmarshal([]byte(s), &data); err != nil {
			return nil, fmt.Errorf("youpin balance: parse data string: %w", err)
		}
	}

	available, _ := strconv.ParseFloat(data.AvailableTotalAmount, 64)
	frozen, _ := strconv.ParseFloat(data.FrozeTotalAmount, 64)
	instant, _ := strconv.ParseFloat(data.TradeOnlyTotalAmount, 64)
	purchase, _ := strconv.ParseFloat(data.PurchaseBalance, 64)

	return &platform.Balance{
		Available: available,
		Frozen:    frozen,
		Instant:   instant,
		Purchase:  purchase,
	}, nil
}

func (c *Client) GetBillHistory(_ context.Context, _ ...platform.QueryOption) ([]platform.BillRecord, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// doRequest serializes the body and delegates to BaseClient (rate limiting is handled by httpclient).
func (c *Client) doRequest(ctx context.Context, method, path string, query map[string]string, reqBody map[string]any) (int, []byte, error) {
	var bodyBytes []byte
	if reqBody != nil {
		bodyBytes, _ = json.Marshal(reqBody)
	}
	return c.DoRequest(ctx, method, path, query, bodyBytes, c.headers())
}

// call makes an HTTP request, checks for errors, and unmarshals the response.
func (c *Client) call(ctx context.Context, method, path string, query map[string]string, reqBody map[string]any, result any) error {
	statusCode, body, err := c.doRequest(ctx, method, path, query, reqBody)
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		return fmt.Errorf("youpin API returned HTTP %d: %s", statusCode, string(body))
	}
	if err := c.checkAPIError(body); err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+strings.TrimSpace(c.token))
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("User-Agent", "okhttp/3.14.9")
	h.Set("App-Version", "5.28.3")
	h.Set("AppType", "4")
	h.Set("deviceType", "1")
	h.Set("package-type", "uuyp")
	h.Set("DeviceToken", c.deviceToken)
	h.Set("DeviceId", c.deviceID)
	h.Set("Gameid", "730")
	h.Set("platform", "android")

	c.ukMu.Lock()
	if c.uk != "" {
		h.Set("uk", c.uk)
	}
	c.ukMu.Unlock()

	deviceInfo := fmt.Sprintf(
		`{"deviceId":"%s","deviceType":"%s","hasSteamApp":1,"requestTag":"%s","systemName":"Android","systemVersion":"15"}`,
		c.deviceID, c.deviceID, strings.ToUpper(randomString(32)),
	)
	h.Set("Device-Info", deviceInfo)
	return h
}

// checkAPIError checks for known YouPin error codes in the response.
func (c *Client) checkAPIError(body []byte) error {
	var resp struct {
		Code  int `json:"Code"`
		CodeL int `json:"code"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil
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
		return nil
	}
	if code != 0 {
		c.Log.Warn("youpin: API returned non-zero code", "code", code)
	}
	return nil
}

func parseWear(s string) float64 {
	if s == "" {
		return 0
	}
	var v float64
	if _, err := fmt.Sscanf(s, "%f", &v); err == nil {
		return v
	}
	return 0
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
