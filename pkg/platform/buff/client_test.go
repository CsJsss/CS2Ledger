package buff

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

func TestClient_FetchBuyHistory(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/market/buy_order/history" {
			page++
			resp := map[string]any{
				"code": "OK",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"state":         "SUCCESS",
							"transact_time": 1736000000,
							"price":         "12.50",
							"fee":           "0",
							"goods_id":      123456,
							"asset_info": map[string]any{
								"assetid":    "asset-1",
								"classid":    "class-1",
								"instanceid": "instance-1",
								"paintwear":  "0.15",
								"info": map[string]any{
									"paintseed":  123,
									"paintindex": 44,
									"stickers":   []any{},
									"keychains":  []any{},
								},
							},
						},
					},
					"goods_infos": map[string]any{
						"123456": map[string]string{"short_name": "AK-47 | Redline"},
					},
					"total_page": 1,
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer srv.Close()

	c := New("test-cookie", logfx.NewNop())
	c.BaseURL = srv.URL
	trades, err := c.GetBuyHistory(context.Background())
	if err != nil {
		t.Fatalf("FetchBuyHistory: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	tr := trades[0]
	if tr.ItemName != "AK-47 | Redline" {
		t.Errorf("expected AK-47 | Redline, got %q", tr.ItemName)
	}
	if tr.TradeType != "buy" {
		t.Errorf("expected buy, got %s", tr.TradeType)
	}
	if tr.UnitPrice != 1250 {
		t.Errorf("expected 1250 (12.50 yuan in fen), got %d", tr.UnitPrice)
	}
}

func TestClient_FetchBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"data": map[string]string{
				"cash_amount":     "100.00",
				"security_amount": "50.00",
				"frozen_amount":   "30.00",
			},
		})
	}))
	defer srv.Close()

	c := New("test-cookie", logfx.NewNop())
	c.BaseURL = srv.URL
	bal, err := c.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("FetchBalance: %v", err)
	}
	if bal.Available != 100.0 {
		t.Errorf("expected available 100.00, got %v", bal.Available)
	}
	if bal.Purchase != 50.0 {
		t.Errorf("expected purchase 50.00, got %v", bal.Purchase)
	}
	if bal.Frozen != 30.0 {
		t.Errorf("expected frozen 30.00, got %v", bal.Frozen)
	}
}

func TestClient_Verify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "OK"})
	}))
	defer srv.Close()

	c := New("test-cookie", logfx.NewNop())
	c.BaseURL = srv.URL
	if err := c.Verify(context.Background()); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestClient_VerifyFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "AUTH_FAILED"})
	}))
	defer srv.Close()

	c := New("bad-cookie", logfx.NewNop())
	c.BaseURL = srv.URL
	if err := c.Verify(context.Background()); err == nil {
		t.Fatal("expected error for invalid credential")
	}
}
