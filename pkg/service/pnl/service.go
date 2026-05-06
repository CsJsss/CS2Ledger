package pnl

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type PnlSummaryView struct {
	TotalTrades  int64 `json:"totalTrades"`
	TotalGrossPl int64 `json:"totalGrossPl"`
	TotalFee     int64 `json:"totalFee"`
	TotalNetPl   int64 `json:"totalNetPl"`
}

type MonthlyPLView struct {
	Month string `json:"month"`
	NetPl int64  `json:"netPl"`
}

type PnlInterface interface {
	GetDaily(accountID uint) ([]model.PnlDaily, error)
	GetSummary(accountID uint) (*PnlSummaryView, error)
	GetMonthlyBreakdown(accountID uint, year int) ([]MonthlyPLView, error)
	ProcessPending(accountID uint) (int, error)
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (s *service) GetDaily(accountID uint) ([]model.PnlDaily, error) {
	return s.orm.FindPnlByAccount(accountID)
}

func (s *service) GetSummary(accountID uint) (*PnlSummaryView, error) {
	daily, err := s.orm.FindPnlByAccount(accountID)
	if err != nil {
		return nil, err
	}
	sum := &PnlSummaryView{}
	for _, d := range daily {
		sum.TotalTrades += d.TradeCount
		sum.TotalGrossPl += d.GrossPl
		sum.TotalFee += d.Fee
		sum.TotalNetPl += d.NetPl
	}
	return sum, nil
}

func (s *service) GetMonthlyBreakdown(accountID uint, year int) ([]MonthlyPLView, error) {
	daily, err := s.orm.FindPnlByAccount(accountID)
	if err != nil {
		return nil, err
	}
	monthMap := make(map[string]int64)
	for _, d := range daily {
		if len(d.Date) >= 7 && d.Date[:4] == itoa(year) {
			month := d.Date[:7]
			monthMap[month] += d.NetPl
		}
	}
	views := make([]MonthlyPLView, 0, len(monthMap))
	for m, pl := range monthMap {
		views = append(views, MonthlyPLView{Month: m, NetPl: pl})
	}
	return views, nil
}

func itoa(i int) string {
	return string(rune('0'+i/1000%10)) + string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
}

func (s *service) ProcessPending(accountID uint) (int, error) {
	sells, err := s.orm.FindUnmatchedSells(accountID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, sell := range sells {
		buys, err := s.orm.FindUnmatchedBuys(accountID, sell.AssetID)
		if err != nil {
			continue
		}

		matched := false
		for _, buy := range buys {
			used, _ := s.orm.CountMatchedSellsForBuy(buy.ID)
			if used > 0 {
				continue
			}

			if err := s.orm.SetMatchedBuy(sell.ID, buy.ID); err != nil {
				continue
			}

			grossPl := (sell.UnitPrice - buy.UnitPrice) * sell.Quantity
			fee := sell.Fee + buy.Fee
			netPl := grossPl - fee

			if err := s.orm.UpsertDailyPnl(accountID, sell.TradeAt, grossPl, fee, netPl); err != nil {
				continue
			}

			_ = s.orm.RemoveInventoryByAssetID(accountID, sell.AssetID)
			matched = true
			break
		}
		if matched {
			count++
		}
	}

	return count, nil
}

var Module = fx.Module("pnl",
	logfx.WithComponent("pnl"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) PnlInterface { return s }, fx.As(new(PnlInterface))),
	),
)
