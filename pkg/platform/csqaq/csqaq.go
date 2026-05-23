package csqaq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

const (
	apiBase  = "https://api.csqaq.com/api/v1"
	priceURL = apiBase + "/goods/getPriceByMarketHashName"
	maxBatch = 50
)

type Client struct {
	apiToken string
	client   *http.Client
}

func New(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		client:   &http.Client{Timeout: 30 * time.Second},
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, priceURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("csqaq: create request: %w", err)
		}
		req.Header.Set("ApiToken", c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("csqaq: request: %w", err)
		}
		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("csqaq: read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("csqaq: http %d: %s", resp.StatusCode, string(respBody))
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

		// Rate-limit: 1s between batches, respect context cancellation
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
