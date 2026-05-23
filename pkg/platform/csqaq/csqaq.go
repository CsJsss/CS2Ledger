package csqaq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/httpclient"
)

type Logger interface {
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Debug(string, ...any) {}
func (nopLogger) Warn(string, ...any)  {}

const (
	apiBase       = "https://api.csqaq.com/api/v1"
	pricePath     = "/goods/getPriceByMarketHashName"
	getGoodIDPath = "/info/get_good_id"
	maxBatch      = 50
)

type Client struct {
	apiToken string
	client   *httpclient.Client
	log      Logger
}

func New(apiToken string, log Logger) *Client {
	if log == nil {
		log = nopLogger{}
	}
	return &Client{
		apiToken: apiToken,
		client: httpclient.New(
			httpclient.WithBaseURL(apiBase),
			httpclient.WithName("csqaq"),
			httpclient.WithNoRetry(),
		),
		log: log,
	}
}

type apiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Success map[string]struct {
			BuffSellPrice  float64 `json:"buffSellPrice"`
			BuffSellNum    int     `json:"buffSellNum"`
			YypSellPrice   float64 `json:"yyypSellPrice"`
			YypSellNum     int     `json:"yyypSellNum"`
			SteamSellPrice float64 `json:"steamSellPrice"`
			SteamSellNum   int     `json:"steamSellNum"`
		} `json:"success"`
		Error []string `json:"error"`
	} `json:"data"`
}

type goodsInfo struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	MarketHashName string `json:"market_hash_name"`
}

type getGoodIDResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Data      map[string]goodsInfo `json:"data"`
		PageIndex int                  `json:"page_index"`
		PageSize  int                  `json:"page_size"`
		Total     int                  `json:"total"`
	} `json:"data"`
}

var exteriorCNToEN = map[string]string{
	"崭新出厂": "Factory New",
	"略有磨损": "Minimal Wear",
	"久经沙场": "Field-Tested",
	"破损不堪": "Well-Worn",
	"战痕累累": "Battle-Scarred",
}

var exteriorENToCN = map[string]string{
	"Factory New":    "崭新出厂",
	"Minimal Wear":   "略有磨损",
	"Field-Tested":   "久经沙场",
	"Well-Worn":      "破损不堪",
	"Battle-Scarred": "战痕累累",
}

// ResolveGoodsInfo queries csqaq's get_good_id API by item name and matches
// the result by exterior. Returns the csqaq goods ID and market_hash_name.
func (c *Client) ResolveGoodsInfo(ctx context.Context, itemName, exterior string) (int, string, error) {
	// Try with full item name first.
	search := itemName
	c.log.Debug("csqaq: resolve goods search", "search", search)
	id, mhn, _ := c.searchGoods(ctx, search, itemName, exterior)
	if id != 0 {
		return id, mhn, nil
	}

	// Fallback: search with skin name + exterior, e.g. "黑色魅影 (久经沙场)".
	if exterior != "" {
		if idx := strings.LastIndex(itemName, "|"); idx >= 0 {
			skin := strings.TrimSpace(itemName[idx+1:])
			search = skin + " (" + exterior + ")"
			c.log.Debug("csqaq: fallback search", "search", search)
			id, mhn, _ = c.searchGoods(ctx, search, itemName, exterior)
			if id != 0 {
				c.log.Debug("csqaq: fallback matched", "goodID", id, "mhn", mhn)
				return id, mhn, nil
			}
		}
	}

	c.log.Debug("csqaq: resolve goods no match", "itemName", itemName, "exterior", exterior)
	return 0, "", nil
}

func (c *Client) searchGoods(ctx context.Context, search, itemName, exterior string) (int, string, error) {
	body, err := json.Marshal(map[string]any{
		"page_index": 1,
		"page_size":  20,
		"search":     search,
	})
	if err != nil {
		return 0, "", fmt.Errorf("csqaq: marshal get_good_id request: %w", err)
	}

	headers := http.Header{}
	headers.Set("ApiToken", c.apiToken)
	headers.Set("Content-Type", "application/json")

	status, respBody, err := c.client.DoRequest(ctx, "POST", getGoodIDPath, nil, body, headers)
	if err != nil {
		return 0, "", fmt.Errorf("csqaq: get_good_id request: %w", err)
	}
	if status != http.StatusOK {
		return 0, "", fmt.Errorf("csqaq: get_good_id http %d: %s", status, string(respBody))
	}

	var ar getGoodIDResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		return 0, "", fmt.Errorf("csqaq: unmarshal get_good_id: %w", err)
	}
	if ar.Code != 200 {
		return 0, "", fmt.Errorf("csqaq: get_good_id api error code=%d msg=%q", ar.Code, ar.Msg)
	}

	id, mhn := matchGoods(ar.Data.Data, itemName, exterior)
	return id, mhn, nil
}

// normalizeName removes spaces and normalizes parentheses for fuzzy matching.
func normalizeName(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "（", "(")
	s = strings.ReplaceAll(s, "）", ")")
	return s
}

func matchGoods(data map[string]goodsInfo, itemName, exterior string) (int, string) {
	if len(data) == 0 {
		return 0, ""
	}

	// Single result — use directly.
	if len(data) == 1 {
		for _, v := range data {
			return v.ID, v.MarketHashName
		}
	}

	normItem := normalizeName(itemName)

	// Multiple results — try to disambiguate by exterior.
	if exterior != "" {
		cnExt := exterior
		if en, ok := exteriorENToCN[exterior]; ok {
			cnExt = en
		}
		enExt := exterior
		if cn, ok := exteriorCNToEN[exterior]; ok {
			enExt = cn
		}

		for _, v := range data {
			name := normalizeName(v.Name)
			if strings.HasSuffix(name, "("+cnExt+")") || strings.HasSuffix(name, "("+enExt+")") {
				return v.ID, v.MarketHashName
			}
		}
	}

	// Fallback: try normalized item name match.
	for _, v := range data {
		if normalizeName(v.Name) == normItem {
			return v.ID, v.MarketHashName
		}
	}

	return 0, ""
}

func (c *Client) GetPrices(ctx context.Context, marketHashNames []string) ([]platform.PriceInfo, error) {
	var all []platform.PriceInfo

	for i := 0; i < len(marketHashNames); i += maxBatch {
		end := i + maxBatch
		if end > len(marketHashNames) {
			end = len(marketHashNames)
		}
		batch := marketHashNames[i:end]

		body, err := json.Marshal(map[string]any{"marketHashNameList": batch})
		if err != nil {
			return nil, fmt.Errorf("csqaq: marshal request: %w", err)
		}

		headers := http.Header{}
		headers.Set("ApiToken", c.apiToken)
		headers.Set("Content-Type", "application/json")

		status, respBody, err := c.client.DoRequest(ctx, "POST", pricePath, nil, body, headers)
		if err != nil {
			return nil, fmt.Errorf("csqaq: request: %w", err)
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("csqaq: http %d: %s", status, string(respBody))
		}

		var ar apiResponse
		if err := json.Unmarshal(respBody, &ar); err != nil {
			return nil, fmt.Errorf("csqaq: unmarshal body=%q: %w", string(respBody), err)
		}
		if ar.Code != 200 {
			return nil, fmt.Errorf("csqaq: api error code=%d msg=%q", ar.Code, ar.Msg)
		}

		for name, item := range ar.Data.Success {
			all = append(all, platform.PriceInfo{
				MarketHashName: name,
				BuffPrice:      item.BuffSellPrice,
				BuffVolume:     item.BuffSellNum,
				YoupinPrice:    item.YypSellPrice,
				YoupinVolume:   item.YypSellNum,
				SteamPrice:     item.SteamSellPrice,
				SteamVolume:    item.SteamSellNum,
			})
		}

		// Rate-limit: 1s between batches
		if end < len(marketHashNames) {
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return all, ctx.Err()
			}
		}
	}

	return all, nil
}
