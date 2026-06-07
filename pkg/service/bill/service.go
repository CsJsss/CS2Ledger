package bill

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type BillFilter struct {
	TypeID    int
	Platform  string
	StartTime int64
	EndTime   int64
}

type BillInterface interface {
	List(accountID uint, page, pageSize int, f BillFilter) (*PaginatedBills, error)
	SumRentalIncome(accountID uint) (int64, error)
	ChartData(accountID uint, f BillFilter) ([]orm.DailyBillSummary, error)
}

type PaginatedBills struct {
	Records    []model.BillRecord `json:"records"`
	TotalCount int64              `json:"totalCount"`
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (s *service) SumRentalIncome(accountID uint) (int64, error) {
	income, err := s.orm.SumBillsByTypes(accountID, []int{
		model.BillTypeRentalIncome,
		model.BillTypeRenewalRental,
	})
	if err != nil {
		return 0, err
	}
	fee, err := s.orm.SumBillsByTypes(accountID, []int{
		model.BillTypeRentalFee,
	})
	if err != nil {
		return 0, err
	}
	return income + fee, nil
}

func (s *service) ChartData(accountID uint, f BillFilter) ([]orm.DailyBillSummary, error) {
	return s.orm.SumBillByDay(accountID, orm.BillFilter{
		TypeID:    f.TypeID,
		Platform:  f.Platform,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	})
}

func (s *service) List(accountID uint, page, pageSize int, f BillFilter) (*PaginatedBills, error) {
	offset := (page - 1) * pageSize
	of := orm.BillFilter{
		TypeID:    f.TypeID,
		Platform:  f.Platform,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}

	var records []model.BillRecord
	var total int64
	var err error

	if accountID == 0 {
		records, err = s.orm.ListAllBills(pageSize, offset, of)
		total, _ = s.orm.CountAllBills(of)
	} else {
		records, err = s.orm.ListBillsByAccount(accountID, pageSize, offset, of)
		total, _ = s.orm.CountBillsByAccount(accountID, of)
	}
	if err != nil {
		return nil, err
	}

	return &PaginatedBills{
		Records:    records,
		TotalCount: total,
	}, nil
}

var Module = fx.Module("bill",
	logfx.WithComponent("bill"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) BillInterface { return s }, fx.As(new(BillInterface))),
	),
)
