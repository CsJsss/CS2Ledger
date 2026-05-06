package buff

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
							"income":        "12.50",
							"goods_id":      123456,
							"asset_info":    map[string]string{"assetid": "asset-1"},
						},
					},
					"goods_infos": map[string]any{
						"123456": map[string]string{"short_name": "AK-47 | Redline"},
					},
					"total_pages": 1,
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer srv.Close()

	c := New("test-cookie")
	c.BaseURL = srv.URL
	trades, err := c.FetchBuyHistory(context.Background(), 0)
	if err != nil {
		t.Fatalf("FetchBuyHistory: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	tr := trades[0]
	if tr.ItemName != "AK-47 | Redline" {
		t.Errorf("expected AK-47 | Redline, got %s", tr.ItemName)
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
				"balance":        "100.00",
				"frozen_balance": "50.00",
			},
		})
	}))
	defer srv.Close()

	c := New("test-cookie")
	c.BaseURL = srv.URL
	bal, err := c.FetchBalance(context.Background())
	if err != nil {
		t.Fatalf("FetchBalance: %v", err)
	}
	if bal.Available != 10000 {
		t.Errorf("expected available 10000, got %d", bal.Available)
	}
	if bal.Purchase != 5000 {
		t.Errorf("expected purchase 5000, got %d", bal.Purchase)
	}
}

func TestClient_Verify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "OK"})
	}))
	defer srv.Close()

	c := New("test-cookie")
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

	c := New("bad-cookie")
	c.BaseURL = srv.URL
	if err := c.Verify(context.Background()); err == nil {
		t.Fatal("expected error for invalid credential")
	}
}
