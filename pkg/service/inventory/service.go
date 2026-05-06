package inventory

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type InventoryInterface interface {
	List(accountID uint, status string) ([]model.InventoryItem, error)
	GetItemDetail(accountID uint, assetID string) (*ItemDetail, error)
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (s *service) List(accountID uint, status string) ([]model.InventoryItem, error) {
	return s.orm.FindInventoryByAccount(accountID, status)
}

func (s *service) GetItemDetail(accountID uint, assetID string) (*ItemDetail, error) {
	item, err := s.orm.FindInventoryByAssetID(accountID, assetID)
	if err != nil || item == nil {
		return nil, err
	}

	rentals, err := s.orm.FindRentalsByAssetID(accountID, assetID)
	if err != nil {
		rentals = nil
	}

	var totalDays, totalIncome int64
	for _, r := range rentals {
		totalDays += r.DurationDays
		totalIncome += r.Income
	}

	return &ItemDetail{
		Item:          *item,
		RentalHistory: rentals,
		RentalSummary: RentalSummary{
			TotalDays:   totalDays,
			TotalIncome: totalIncome,
			RentCount:   len(rentals),
		},
	}, nil
}

type ItemDetail struct {
	Item          model.InventoryItem  `json:"item"`
	RentalHistory []model.RentalRecord `json:"rentalHistory"`
	RentalSummary RentalSummary        `json:"rentalSummary"`
}

type RentalSummary struct {
	TotalDays   int64 `json:"totalDays"`
	TotalIncome int64 `json:"totalIncome"`
	RentCount   int   `json:"rentCount"`
}

var Module = fx.Module("inventory",
	logfx.WithComponent("inventory"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) InventoryInterface { return s }, fx.As(new(InventoryInterface))),
	),
)
