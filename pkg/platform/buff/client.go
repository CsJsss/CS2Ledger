package buff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils"
)

type Client struct {
	platform.BaseClient
	cookie string
}

func New(cookie string) *Client {
	c := &Client{
		BaseClient: platform.NewBaseClient(utils.PlatformBuff, "https://buff.163.com"),
		cookie:     cookie,
	}
	c.init()
	return c
}

// init primes the session cookie and validates connectivity.
func (c *Client) init() {
	req, _ := http.NewRequest("GET", c.BaseURL+"/api/message/notification", nil)
	req.Header.Set("User-Agent", platform.RandomUA())
	req.Header.Set("Cookie", c.cookie)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		c.Log.Warn("buff: init notification call failed", "err", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	c.Log.Debug("buff: init notification", "status", resp.StatusCode, "body", string(body))
}

func (c *Client) Verify(ctx context.Context) error {
	c.Log.Info("buff: verifying")
	resp, err := c.doGet(ctx, "/account/api/user/info")
	if err != nil {
		c.Log.Warn("buff: verify failed", "err", err)
		return fmt.Errorf("buff verify: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("buff verify: %w", err)
	}
	if result.Code != "OK" {
		c.Log.Warn("buff: verify invalid credential", "code", result.Code)
		return fmt.Errorf("buff verify: credential invalid (code=%s)", result.Code)
	}
	c.Log.Info("buff: verify ok")
	return nil
}

func (c *Client) FetchBuyHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Info("buff: fetching buy history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "buy", 2*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchBuyPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("buff: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	path := fmt.Sprintf("/api/market/buy_order/history?game=csgo&page_num=%d&page_size=100", page)
	resp, err := c.doGet(ctx, path)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result buyOrderHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, err
	}
	if result.Code != "OK" {
		return nil, false, fmt.Errorf("buff API error: code=%s", result.Code)
	}

	sinceSec := since / 1000
	trades := make([]platform.TradeRecord, 0, len(result.Data.Items))
	for _, item := range result.Data.Items {
		if item.TransactTime < sinceSec {
			continue
		}
		if item.State != "SUCCESS" {
			continue
		}
		price, _ := strconv.ParseFloat(item.Income, 64)
		goodsID := strconv.FormatInt(item.GoodsID, 10)
		itemName := goodsID
		if info, ok := result.Data.GoodsInfos[goodsID]; ok {
			itemName = info.ShortName
		}
		trades = append(trades, platform.TradeRecord{
			ExternalID: fmt.Sprintf("buff-buy-%s", item.AssetInfo.AssetID),
			AssetID:    goodsID,
			ItemName:   itemName,
			TradeType:  "buy",
			Quantity:   1,
			UnitPrice:  int64(price * 100),
			TotalPrice: int64(price * 100),
			Fee:        0,
			TradeAt:    item.TransactTime * 1000,
		})
	}
	hasMore := page < result.Data.TotalPages
	if result.Data.TotalPages == 0 {
		if result.Data.Total > 0 {
			hasMore = page*100 < result.Data.Total
		} else {
			hasMore = len(result.Data.Items) == 100
		}
	}
	c.Log.Debug("buff: buy page", "page", page, "items", len(result.Data.Items), "total_pages", result.Data.TotalPages, "total", result.Data.Total, "has_more", hasMore)
	return trades, hasMore, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64) ([]platform.TradeRecord, bool, error) {
	path := fmt.Sprintf("/api/market/sell_order/history?appid=730&mode=1&page_num=%d&page_size=100", page)
	resp, err := c.doGet(ctx, path)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result sellOrderHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, err
	}
	if result.Code != "OK" {
		return nil, false, fmt.Errorf("buff API error: code=%s", result.Code)
	}

	sinceSec := since / 1000
	trades := make([]platform.TradeRecord, 0, len(result.Data.Items))
	for _, item := range result.Data.Items {
		if item.CreateTime < sinceSec {
			continue
		}
		if item.Status != "deal" {
			continue
		}
		price, _ := strconv.ParseFloat(item.Price, 64)
		goodsID := strconv.FormatInt(item.GoodsID, 10)
		itemName := goodsID
		if info, ok := result.Data.GoodsInfos[goodsID]; ok {
			itemName = info.ShortName
		}
		trades = append(trades, platform.TradeRecord{
			ExternalID: fmt.Sprintf("buff-sell-%s", item.ID),
			AssetID:    goodsID,
			ItemName:   itemName,
			TradeType:  "sell",
			Quantity:   1,
			UnitPrice:  int64(price * 100),
			TotalPrice: int64(price * 100),
			Fee:        0,
			TradeAt:    item.CreateTime * 1000,
		})
	}
	hasMore := page < result.Data.TotalPages
	if result.Data.TotalPages == 0 {
		if result.Data.Total > 0 {
			hasMore = page*100 < result.Data.Total
		} else {
			hasMore = len(result.Data.Items) == 100
		}
	}
	c.Log.Debug("buff: sell page", "page", page, "items", len(result.Data.Items), "total_pages", result.Data.TotalPages, "total", result.Data.Total, "has_more", hasMore)
	return trades, hasMore, nil
}

func (c *Client) FetchSellHistory(ctx context.Context, since int64) ([]platform.TradeRecord, error) {
	c.Log.Info("buff: fetching sell history", "since", since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, "sell", 2*time.Second,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			return c.fetchSellPage(ctx, page, since)
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("buff: sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) FetchBalance(ctx context.Context) (*platform.Balance, error) {
	c.Log.Info("buff: fetching balance")
	resp, err := c.doGet(ctx, "/api/asset/get_brief_asset")
	if err != nil {
		return nil, fmt.Errorf("buff balance: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != "OK" {
		return nil, fmt.Errorf("buff balance: code=%s", result.Code)
	}

	available, _ := strconv.ParseFloat(result.Data.Balance, 64)
	purchase, _ := strconv.ParseFloat(result.Data.FrozenBalance, 64)

	return &platform.Balance{
		Available: int64(available * 100),
		Purchase:  int64(purchase * 100),
	}, nil
}

func (c *Client) doGet(ctx context.Context, path string) (*http.Response, error) {
	var lastResp *http.Response
	for i := 0; i < 10; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", platform.RandomUA())
		req.Header.Set("Cookie", c.cookie)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}

		// Retry on "系统繁忙" (server busy).
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))

		if strings.Contains(string(body), "系统繁忙") {
			c.Log.Warn("buff: 系统繁忙 retrying", "attempt", i+1)
			time.Sleep(2 * time.Second)
			lastResp = resp
			continue
		}
		return resp, nil
	}
	// All retries exhausted, return last response.
	if lastResp != nil {
		return lastResp, nil
	}
	return nil, fmt.Errorf("buff: request failed after 10 retries")
}
