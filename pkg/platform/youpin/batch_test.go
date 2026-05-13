package youpin

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

func TestFetchBuyBatch_Real(t *testing.T) {
	token := os.Getenv("YOUPIN_TOKEN")
	if token == "" {
		t.Skip("YOUPIN_TOKEN not set")
	}

	c := New(token, logfx.NewNop())
	if err := c.Verify(context.Background()); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	t.Logf("userID: %d", c.userID)

	// Fill these in with values you want to test.
	orderNo := os.Getenv("YOUPIN_ORDER_NO")
	buyerUserID := c.userID
	if orderNo == "" {
		t.Skip("YOUPIN_ORDER_NO not set")
	}

	t.Logf("calling fetchBuyBatch(orderNo=%s, buyerUserID=%d)", orderNo, buyerUserID)

	// Fetch raw response first for inspection.
	body := map[string]any{
		"orderNo":   orderNo,
		"userId":    buyerUserID,
		"Sessionid": c.deviceID,
	}
	_, respBody, err := c.doRequest(context.Background(), "POST", "/api/youpin/bff/trade/v1/order/query/detail", nil, body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var prettyJSON map[string]any
	if err := json.Unmarshal(respBody, &prettyJSON); err == nil {
		raw, _ := json.MarshalIndent(prettyJSON, "", "  ")
		t.Logf("raw response:\n%s", string(raw))
	}

	trades, err := c.fetchBuyBatch(context.Background(), orderNo, buyerUserID, 0)
	if err != nil {
		t.Fatalf("fetchBuyBatch: %v", err)
	}

	t.Logf("got %d trades from batch order", len(trades))
	for i, tr := range trades {
		t.Logf("[%d] item=%s | ext=%s | unit=%.2f元 | total=%.2f元 | wear=%.4f | paintSeed=%d | paintIndex=%d | tradeAt=%d",
			i,
			tr.ItemName,
			tr.Exterior,
			float64(tr.UnitPrice)/100,
			float64(tr.TotalPrice)/100,
			tr.PaintWear,
			tr.PaintSeed,
			tr.PaintIndex,
			tr.TradeAt,
		)
	}
}
