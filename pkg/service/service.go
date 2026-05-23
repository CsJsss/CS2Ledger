package service

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/service/account"
	"github.com/CsJsss/CS2Ledger/pkg/service/inventory"
	"github.com/CsJsss/CS2Ledger/pkg/service/market"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/service/rental"
	"github.com/CsJsss/CS2Ledger/pkg/service/trade"
	"github.com/CsJsss/CS2Ledger/pkg/utils/configfx"
)

type Service struct {
	acc account.AccountInterface
	trd trade.TradeInterface
	inv inventory.InventoryInterface
	p   pnl.PnlInterface
	rnt rental.RentalInterface
	mkt market.MarketInterface
	cfg configfx.Config
}

func New(
	acc account.AccountInterface,
	trd trade.TradeInterface,
	inv inventory.InventoryInterface,
	p pnl.PnlInterface,
	rnt rental.RentalInterface,
	mkt market.MarketInterface,
	cfg configfx.Config,
) *Service {
	s := &Service{
		acc: acc,
		trd: trd,
		inv: inv,
		p:   p,
		rnt: rnt,
		mkt: mkt,
		cfg: cfg,
	}
	inv.SetPriceProvider(mkt)
	inv.SetPriceSource(cfg.PriceSource)
	return s
}

func (s *Service) Account() account.AccountInterface       { return s.acc }
func (s *Service) Trade() trade.TradeInterface             { return s.trd }
func (s *Service) Inventory() inventory.InventoryInterface { return s.inv }
func (s *Service) Pnl() pnl.PnlInterface                   { return s.p }
func (s *Service) Rental() rental.RentalInterface          { return s.rnt }
func (s *Service) Market() market.MarketInterface          { return s.mkt }
func (s *Service) Config() *configfx.Config                { return &s.cfg }

var Module = fx.Module("service", fx.Provide(New))
