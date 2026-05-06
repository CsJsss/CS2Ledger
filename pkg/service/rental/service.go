package rental

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type RentalInterface interface {
	ListByAsset(accountID uint, assetID string) ([]model.RentalRecord, error)
	ListByAccount(accountID uint) ([]model.RentalRecord, error)
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (s *service) ListByAsset(accountID uint, assetID string) ([]model.RentalRecord, error) {
	return s.orm.FindRentalsByAssetID(accountID, assetID)
}

func (s *service) ListByAccount(accountID uint) ([]model.RentalRecord, error) {
	return s.orm.FindRentalsByAccount(accountID)
}

var Module = fx.Module("rental",
	logfx.WithComponent("rental"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) RentalInterface { return s }, fx.As(new(RentalInterface))),
	),
)
