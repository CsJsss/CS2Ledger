package service

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/service/account"
	"github.com/CsJsss/CS2Ledger/pkg/service/inventory"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/service/rental"
	"github.com/CsJsss/CS2Ledger/pkg/service/trade"
)

type Service struct {
	acc account.AccountInterface
	trd trade.TradeInterface
	inv inventory.InventoryInterface
	p   pnl.PnlInterface
	rnt rental.RentalInterface
}

func New(
	acc account.AccountInterface,
	trd trade.TradeInterface,
	inv inventory.InventoryInterface,
	p pnl.PnlInterface,
	rnt rental.RentalInterface,
) *Service {
	return &Service{
		acc: acc,
		trd: trd,
		inv: inv,
		p:   p,
		rnt: rnt,
	}
}

func (s *Service) Account() account.AccountInterface       { return s.acc }
func (s *Service) Trade() trade.TradeInterface             { return s.trd }
func (s *Service) Inventory() inventory.InventoryInterface { return s.inv }
func (s *Service) Pnl() pnl.PnlInterface                   { return s.p }
func (s *Service) Rental() rental.RentalInterface          { return s.rnt }

var Module = fx.Module("service", fx.Provide(New))
