package platform

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func nopLogger() *logfx.Logger { return logfx.NewNop() }

// ---------------------------------------------------------------------------
// QueryOption / ApplyQueryOpts
// ---------------------------------------------------------------------------

func TestApplyQueryOpts_Defaults(t *testing.T) {
	cfg := ApplyQueryOpts(nil)
	assert.Equal(t, int64(0), cfg.Since)
	assert.Equal(t, 0, cfg.Limit)
	assert.Equal(t, TradeState(""), cfg.TradeState)
	assert.Nil(t, cfg.ExtraParams)
}

func TestApplyQueryOpts_AllOptions(t *testing.T) {
	cfg := ApplyQueryOpts([]QueryOption{
		WithSince(1000),
		WithLimit(50),
		WithTradeState(TradeStateCompleted),
		WithExtraParams(map[string]string{"k": "v"}),
	})
	assert.Equal(t, int64(1000), cfg.Since)
	assert.Equal(t, 50, cfg.Limit)
	assert.Equal(t, TradeStateCompleted, cfg.TradeState)
	assert.Equal(t, map[string]string{"k": "v"}, cfg.ExtraParams)
}

func TestApplyQueryOpts_LaterOptionWins(t *testing.T) {
	cfg := ApplyQueryOpts([]QueryOption{
		WithSince(1000),
		WithSince(2000),
	})
	assert.Equal(t, int64(2000), cfg.Since)
}

// ---------------------------------------------------------------------------
// NormalizeItemName
// ---------------------------------------------------------------------------

func TestNormalizeItemName(t *testing.T) {
	tests := []struct {
		input        string
		wantName     string
		wantExterior string
	}{
		{"蝴蝶刀（★） | 澄澈之水 (久经沙场)", "蝴蝶刀（★） | 澄澈之水", "久经沙场"},
		{"AK-47 | Redline (Field-Tested)", "AK-47 | Redline", "Field-Tested"},
		{"M4A4 | 咆哮 (崭新出厂)", "M4A4 | 咆哮", "崭新出厂"},
		{"AWP | Dragon Lore (Factory New)", "AWP | Dragon Lore", "Factory New"},
		{"Knife (★) | Doppler (略有磨损)", "Knife (★) | Doppler", "略有磨损"},
		{"无磨损皮肤", "无磨损皮肤", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		name, ext := NormalizeItemName(tt.input)
		assert.Equal(t, tt.wantName, name, "name mismatch for %q", tt.input)
		assert.Equal(t, tt.wantExterior, ext, "exterior mismatch for %q", tt.input)
	}
}

// ---------------------------------------------------------------------------
// MockClient (generated mock smoke test)
// ---------------------------------------------------------------------------

func TestMockClient_Verify(t *testing.T) {
	m := NewMockClient(t)
	m.EXPECT().Verify(mock.Anything).Return(nil)

	err := m.Verify(context.Background())
	assert.NoError(t, err)
}

func TestMockClient_Verify_Error(t *testing.T) {
	m := NewMockClient(t)
	m.EXPECT().Verify(mock.Anything).Return(errors.New("auth failed"))

	err := m.Verify(context.Background())
	assert.EqualError(t, err, "auth failed")
}

func TestMockClient_GetBalance(t *testing.T) {
	m := NewMockClient(t)
	expected := &Balance{Available: 100.0, Frozen: 50.0}
	m.EXPECT().GetBalance(mock.Anything).Return(expected, nil)

	b, err := m.GetBalance(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 100.0, b.Available)
	assert.Equal(t, 50.0, b.Frozen)
}

func TestMockClient_GetBuyHistory_MatchOpts(t *testing.T) {
	m := NewMockClient(t)
	m.EXPECT().GetBuyHistory(mock.Anything, mock.MatchedBy(func(opts []QueryOption) bool {
		cfg := ApplyQueryOpts(opts)
		return cfg.Since == 42 && cfg.Limit == 10
	})).Return([]TradeRecord{{ExternalID: "t1"}}, nil)

	records, err := m.GetBuyHistory(context.Background(), WithSince(42), WithLimit(10))
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "t1", records[0].ExternalID)
}

// ---------------------------------------------------------------------------
// FetchAllPages
// ---------------------------------------------------------------------------

func TestFetchAllPages_Empty(t *testing.T) {
	items, err := FetchAllPages(context.Background(), nopLogger(), "test", "buy", 0, 0,
		func(_ context.Context, page int) ([]string, bool, error) {
			return nil, false, nil
		},
	)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFetchAllPages_SinglePage(t *testing.T) {
	items, err := FetchAllPages(context.Background(), nopLogger(), "test", "buy", 0, 0,
		func(_ context.Context, page int) ([]string, bool, error) {
			if page == 1 {
				return []string{"a", "b"}, false, nil
			}
			return nil, false, nil
		},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, items)
}

func TestFetchAllPages_MultiplePages(t *testing.T) {
	pageCalls := 0
	items, err := FetchAllPages(context.Background(), nopLogger(), "t", "d", 0, 0,
		func(_ context.Context, page int) ([]string, bool, error) {
			pageCalls++
			if page <= 3 {
				return []string{"x"}, true, nil
			}
			return nil, false, nil
		},
	)
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, 4, pageCalls)
}

func TestFetchAllPages_Limit(t *testing.T) {
	items, err := FetchAllPages(context.Background(), nopLogger(), "t", "d", 0, 2,
		func(_ context.Context, page int) ([]string, bool, error) {
			return []string{"a", "b", "c"}, true, nil
		},
	)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestFetchAllPages_Error(t *testing.T) {
	sentinel := errors.New("boom")
	items, err := FetchAllPages(context.Background(), nopLogger(), "t", "d", 0, 0,
		func(_ context.Context, page int) ([]string, bool, error) {
			return nil, false, sentinel
		},
	)
	assert.ErrorContains(t, err, "boom")
	assert.Nil(t, items)
}

func TestFetchAllPages_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	items, err := FetchAllPages(ctx, nopLogger(), "t", "d", 0, 0,
		func(_ context.Context, page int) ([]string, bool, error) {
			return nil, true, nil
		},
	)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, items)
}

// ---------------------------------------------------------------------------
// FetchByTimeWindows
// ---------------------------------------------------------------------------

func TestFetchByTimeWindows_SingleWindow(t *testing.T) {
	cfg := QueryConfig{Limit: 1}
	items, err := FetchByTimeWindows(context.Background(), nopLogger(), "eco", "bill", cfg, 30,
		func(_ context.Context, page int, _, _ time.Time) ([]string, bool, error) {
			return []string{"a", "b"}, true, nil
		},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"a"}, items)
}

func TestFetchByTimeWindows_ConsecutiveEmptyBreaks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var calls int
	_, _ = FetchByTimeWindows(ctx, nopLogger(), "eco", "bill", QueryConfig{}, 30,
		func(_ context.Context, page int, _, _ time.Time) ([]string, bool, error) {
			calls++
			return nil, false, nil
		},
	)
	assert.GreaterOrEqual(t, calls, 12)
}

func TestFetchByTimeWindows_SinceTime_Respected(t *testing.T) {
	since := time.Now().Add(-1 * time.Hour).UnixMilli()
	cfg := QueryConfig{Since: since}
	var seenWindows []time.Time
	_, _ = FetchByTimeWindows(context.Background(), nopLogger(), "eco", "bill", cfg, 30,
		func(_ context.Context, page int, ws, _ time.Time) ([]string, bool, error) {
			seenWindows = append(seenWindows, ws)
			return nil, false, nil
		},
	)
	assert.LessOrEqual(t, len(seenWindows), 3)
	for _, ws := range seenWindows {
		assert.False(t, ws.Before(time.UnixMilli(since)), "window start %v before since %v", ws, time.UnixMilli(since))
	}
}
