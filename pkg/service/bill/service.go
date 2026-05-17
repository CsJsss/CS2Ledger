package bill

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type BillInterface interface {
	List(accountID uint) ([]model.BillRecord, error)
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (s *service) List(accountID uint) ([]model.BillRecord, error) {
	if accountID == 0 {
		return s.orm.ListAllBills(1000, 0)
	}
	return s.orm.ListBillsByAccount(accountID, 1000, 0)
}

var Module = fx.Module("bill",
	logfx.WithComponent("bill"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) BillInterface { return s }, fx.As(new(BillInterface))),
	),
)
