package youpin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_FetchBuyHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"code": 0,
			"data": map[string]any{
				"orderList": []map[string]any{
					{
						"id":              "1001",
						"orderId":         "1001",
						"orderStatusName": "已完成",
						"finishOrderTime": "1736000000000",
						"commodityNum":    "1",
						"productDetailList": []map[string]any{
							{
								"assertId":      "123456",
								"commodityId":   "0",
								"commodityName": "AK-47 | Redline",
								"price":         "125000", // price in fen
							},
						},
					},
				},
				"total": 1,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New("test-token")
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
	// 125000 fen
	if tr.UnitPrice != 125000 {
		t.Errorf("expected 125000 fen, got %v", tr.UnitPrice)
	}
	if tr.AssetID != "123456" {
		t.Errorf("expected assetID 123456, got %s", tr.AssetID)
	}
}

func TestClient_FetchBuyHistory_SkipsNonCompleted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"code": 0,
			"data": map[string]any{
				"orderList": []map[string]any{
					{
						"id":              "1001",
						"orderId":         "1001",
						"orderStatusName": "交易中",
						"finishOrderTime": "1736000000000",
						"commodityNum":    "1",
						"productDetailList": []map[string]any{
							{
								"assertId":      "123456",
								"commodityId":   "0",
								"commodityName": "M4A1-S | Guardian",
								"price":         "50000",
							},
						},
					},
				},
				"total": 1,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New("test-token")
	c.BaseURL = srv.URL
	trades, err := c.FetchBuyHistory(context.Background(), 0)
	if err != nil {
		t.Fatalf("FetchBuyHistory: %v", err)
	}
	if len(trades) != 0 {
		t.Fatalf("expected 0 trades (non-completed), got %d", len(trades))
	}
}

func TestClient_FetchBuyHistory_AssetIDFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"code": 0,
			"data": map[string]any{
				"orderList": []map[string]any{
					{
						"id":              "1001",
						"orderId":         "1001",
						"orderStatusName": "已完成",
						"finishOrderTime": "1736000000000",
						"commodityNum":    "1",
						"productDetailList": []map[string]any{
							{
								"assertId":      "0",
								"commodityId":   "99999",
								"commodityName": "AWP | Dragon Lore",
								"price":         "10000000",
							},
						},
					},
				},
				"total": 1,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New("test-token")
	c.BaseURL = srv.URL
	trades, err := c.FetchBuyHistory(context.Background(), 0)
	if err != nil {
		t.Fatalf("FetchBuyHistory: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].AssetID != "99999" {
		t.Errorf("expected fallback assetID 99999, got %s", trades[0].AssetID)
	}
}

func TestClient_Verify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Code": 0, "Data": map[string]any{"UserId": 12345}})
	}))
	defer srv.Close()

	c := New("test-token")
	c.BaseURL = srv.URL
	if err := c.Verify(context.Background()); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if c.userID != 12345 {
		t.Errorf("expected userID 12345, got %d", c.userID)
	}
}

func TestClient_VerifyFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]int{"Code": 401})
	}))
	defer srv.Close()

	c := New("bad-token")
	c.BaseURL = srv.URL
	if err := c.Verify(context.Background()); err == nil {
		t.Fatal("expected error for invalid token")
	}
}
