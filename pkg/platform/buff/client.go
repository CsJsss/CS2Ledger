package buff

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

const (
	StatusSuccess   = "SUCCESS"
	DefaultPageSize = "100"
)

type Client struct {
	platform.BaseClient
	cookie   string
	initOnce sync.Once
}

func New(cookie string, logger *logfx.Logger) *Client {
	return &Client{
		BaseClient: platform.NewBaseClient(platform.PlatformBuff, "https://buff.163.com", logger),
		cookie:     cookie,
	}
}

// primeSession sends a priming request to establish a session cookie.
func (c *Client) primeSession() {
	c.initOnce.Do(func() {
		req, _ := http.NewRequest("GET", c.BaseURL+"/api/message/notification", nil)
		req.Header.Set("User-Agent", platform.RandomUA())
		req.Header.Set("Cookie", c.cookie)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.Log.Warn("buff: init notification call failed", "err", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		c.Log.Debug("buff: init notification", "status", resp.StatusCode, "body", string(body))
	})
}

func (c *Client) Verify(ctx context.Context) error {
	c.primeSession()
	c.Log.DebugContext(ctx, "verifying")
	_, body, err := c.doRequest(ctx, "GET", "/account/api/user/info", nil, nil)
	if err != nil {
		c.Log.Warn("verify failed", "err", err)
		return err
	}

	var result userInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if result.Code != "OK" {
		c.Log.Warn("verify invalid credential", "code", result.Code)
		return fmt.Errorf("credential invalid (code=%s)", result.Code)
	}
	c.Log.DebugContext(ctx, "verify ok")
	return nil
}

func (c *Client) GetBuyHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	c.primeSession()
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("fetching buy history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, model.DirectionBuy, 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			items, hasMore, _, err := c.fetchBuyPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams)
			return items, hasMore, err
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("buff: buy history done", "total", len(trades))
	return trades, nil
}

func (c *Client) fetchBuyPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string) ([]platform.TradeRecord, bool, int, error) {
	query := map[string]string{"game": "csgo", "page_num": strconv.Itoa(page), "page_size": DefaultPageSize}
	for k, v := range extra {
		query[k] = v
	}
	_, body, err := c.doRequest(ctx, "GET", "/api/market/buy_order/history", query, nil)
	if err != nil {
		return nil, false, 0, err
	}

	var result buyOrderHistoryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, false, 0, err
	}
	if result.Code != "OK" {
		return nil, false, 0, fmt.Errorf("API error: code=%s", result.Code)
	}

	sinceSec := since / 1000
	trades := make([]platform.TradeRecord, 0, len(result.Data.Items))
	finished := false
	for _, item := range result.Data.Items {
		if item.TransactTime < sinceSec {
			finished = true
			continue
		}
		if tradeState == platform.TradeStateCompleted && item.State != StatusSuccess {
			continue
		}
		trades = append(trades, toBuyTrade(item, result.Data.GoodsInfos))
	}
	c.Log.Debug("buy history", "page", page, "items", len(trades), "total_pages", result.Data.TotalPages, "total", result.Data.Total)
	if len(result.Data.Items) == 0 || finished || (result.Data.TotalPages > 0 && page >= result.Data.TotalPages) {
		return trades, false, result.Data.TotalPages, nil
	}
	return trades, true, result.Data.TotalPages, nil
}

func (c *Client) fetchSellPage(ctx context.Context, page int, since int64, tradeState platform.TradeState, extra map[string]string) ([]platform.TradeRecord, bool, int, error) {
	query := map[string]string{"appid": "730", "mode": "1", "page_num": strconv.Itoa(page), "page_size": DefaultPageSize}
	for k, v := range extra {
		query[k] = v
	}
	_, body, err := c.doRequest(ctx, "GET", "/api/market/sell_order/history", query, nil)
	if err != nil {
		return nil, false, 0, err
	}

	var result sellOrderHistoryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, false, 0, err
	}
	if result.Code != "OK" {
		return nil, false, 0, fmt.Errorf("buff API error: code=%s", result.Code)
	}

	sinceSec := since / 1000
	trades := make([]platform.TradeRecord, 0, len(result.Data.Items))
	finished := false
	for _, item := range result.Data.Items {
		if item.CreateTime < sinceSec {
			finished = true
			continue
		}
		if tradeState == platform.TradeStateCompleted && item.State != StatusSuccess {
			continue
		}
		trades = append(trades, toSellTrade(item, result.Data.GoodsInfos))
	}
	c.Log.Debug("sell history", "page", page, "items", len(trades), "total_pages", result.Data.TotalPages, "total", result.Data.Total)
	if len(result.Data.Items) == 0 || finished || (result.Data.TotalPages > 0 && page >= result.Data.TotalPages) {
		return trades, false, result.Data.TotalPages, nil
	}
	return trades, true, result.Data.TotalPages, nil
}

func (c *Client) GetSellHistory(ctx context.Context, opts ...platform.QueryOption) ([]platform.TradeRecord, error) {
	c.primeSession()
	cfg := platform.ApplyQueryOpts(opts)
	c.Log.Info("fetching sell history", "since", cfg.Since)
	trades, err := platform.FetchAllPages(ctx, c.Log, c.Name, model.DirectionSell, 1*time.Second, cfg.Limit,
		func(ctx context.Context, page int) ([]platform.TradeRecord, bool, error) {
			items, hasMore, _, err := c.fetchSellPage(ctx, page, cfg.Since, cfg.TradeState, cfg.ExtraParams)
			return items, hasMore, err
		},
	)
	if err != nil {
		return trades, err
	}
	c.Log.Info("sell history done", "total", len(trades))
	return trades, nil
}

func (c *Client) GetBalance(ctx context.Context) (*platform.Balance, error) {
	c.primeSession()
	c.Log.Info("fetching balance")
	_, body, err := c.doRequest(ctx, "GET", "/api/asset/get_brief_asset", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("buff balance: %w", err)
	}

	var result balanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Code != "OK" {
		return nil, fmt.Errorf("buff balance: code=%s", result.Code)
	}

	available, _ := strconv.ParseFloat(result.Data.CashAmount, 64)
	purchase, _ := strconv.ParseFloat(result.Data.SecurityAmount, 64)
	frozen, _ := strconv.ParseFloat(result.Data.FrozenAmount, 64)

	return &platform.Balance{
		Available: available,
		Purchase:  purchase,
		Frozen:    frozen,
		Instant:   0,
	}, nil
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("User-Agent", platform.RandomUA())
	h.Set("Cookie", c.cookie)
	h.Set("X-Requested-With", "XMLHttpRequest")
	return h
}

//nolint:unparam // unified signature — method is always GET for buff, kept for consistency
func (c *Client) doRequest(ctx context.Context, method, path string, query map[string]string, body map[string]any) (int, []byte, error) {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	for i := 0; i < 10; i++ {
		statusCode, respBody, err := c.DoRequest(ctx, method, path, query, bodyBytes, c.headers())
		if err != nil {
			return 0, nil, err
		}
		if !strings.Contains(string(respBody), "系统繁忙") {
			return statusCode, respBody, nil
		}
		c.Log.Warn("系统繁忙 retrying", "attempt", i+1)
		time.Sleep(2 * time.Second)
	}
	return 0, nil, fmt.Errorf("buff: request failed after 10 retries")
}

func toCS2Item(ai assetInfo, goodsID int64, goods *goodInfo) model.CS2Item {
	item := model.CS2Item{
		AssetID:              ai.AssetID,
		ClassID:              ai.ClassID,
		InstanceID:           ai.InstanceID,
		GoodsID:              int(goodsID),
		PaintSeed:            ai.Info.PaintSeed,
		PaintIndex:           ai.Info.PaintIndex,
		TradableUnfrozenTime: ai.TradableUnfrozenTime,
	}
	if w, err := strconv.ParseFloat(ai.PaintWear, 64); err == nil {
		item.PaintWear = w
	}
	if ai.Info.IconURL != "" {
		item.IconURL = ai.Info.IconURL
	} else {
		item.IconURL = ai.Info.OriginalIconURL
	}

	for _, s := range ai.Info.Stickers {
		sticker := model.Sticker{
			StickerID: s.StickerID,
			Slot:      s.Slot,
			Wear:      s.Wear,
			Name:      s.Name,
			ImageURL:  s.ImageURL,
			OffsetX:   s.OffsetX,
			OffsetY:   s.OffsetY,
		}
		if rp, err := strconv.ParseFloat(s.ReferencePrice, 64); err == nil {
			sticker.ReferencePrice = rp
		}
		item.Stickers = append(item.Stickers, sticker)
	}
	for _, k := range ai.Info.Keychains {
		item.Keychains = append(item.Keychains, model.Keychain{
			Name:     k.Name,
			ImageURL: k.ImageURL,
		})
	}

	if goods != nil {
		item.MarketHashName = goods.MarketHashName
		name, ext := platform.NormalizeItemName(firstNonEmpty(goods.ShortName, goods.Name))
		item.ItemName = name
		item.IconURL = firstNonEmpty(goods.IconURL, item.IconURL)
		tagExterior := goods.Tags.Exterior.LocalizedName
		item.Exterior = firstNonEmpty(tagExterior, ext)
		item.CategoryGroup = goods.Tags.CategoryGroup.LocalizedName
		item.Rarity = goods.Tags.Rarity.LocalizedName
		item.WeaponName = goods.Tags.Category.LocalizedName
		item.WeaponType = goods.Tags.Type.LocalizedName
		item.Quality = goods.Tags.Quality.LocalizedName
		item.Series = goods.Tags.Series.LocalizedName
		item.Itemset = goods.Tags.Itemset.LocalizedName
		item.WeaponCase = goods.Tags.WeaponCase.LocalizedName
		item.Custom = goods.Tags.Custom.LocalizedName
	}

	return item
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
