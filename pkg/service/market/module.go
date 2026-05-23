package market

import (
	"context"

	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type PriceConfig struct {
	PriceSource string
	CacheTTLMin int
}

type MarketInterface interface {
	GetAllPrices() ([]platform.PriceInfo, error)
	SetProvider(p platform.PriceProvider)
	SetConfig(cfg PriceConfig)
	StartAutoRefresh(ctx context.Context)
	EnsureProvider()
}

var Module = fx.Module("market",
	logfx.WithComponent("market"),
	fx.Provide(
		NewMarketService,
		fx.Annotate(func(s *MarketService) MarketInterface { return s }, fx.As(new(MarketInterface))),
	),
)
