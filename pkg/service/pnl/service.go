package pnl

import (
	"time"

	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type PnlSummaryView struct {
	TotalTrades       int64 `json:"totalTrades"`
	TotalGrossPl      int64 `json:"totalGrossPl"`
	TotalFee          int64 `json:"totalFee"`
	TotalNetPl        int64 `json:"totalNetPl"`
	WithdrawalFee     int64 `json:"withdrawalFee"`
	WithdrawalFeeRate int64 `json:"withdrawalFeeRate"`
}

type MonthlyPLView struct {
	Month         string `json:"month"`
	NetPl         int64  `json:"netPl"`
	WithdrawalFee int64  `json:"withdrawalFee"`
}

type PnlInterface interface {
	GetDaily(accountID uint) ([]model.PnlDaily, error)
	GetSummary(accountID uint) (*PnlSummaryView, error)
	GetMonthlyBreakdown(accountID uint, year int) ([]MonthlyPLView, error)
	RunMatching() (int, error)
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
		if year != 0 && (len(d.Date) < 7 || d.Date[:4] != itoa(year)) {
			continue
		}
		if len(d.Date) >= 7 {
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

// RunMatching clears all existing matches, then re-matches all sells globally
// using FIFO by sell time. Buy must occur before sell (time constraint).
// Matches are quantity-aware: a sell of 15 units may consume multiple buys.
// P&L is attributed to the sell-side account.
func (s *service) RunMatching() (int, error) {
	if err := s.orm.ClearAllMatches(); err != nil {
		return 0, err
	}

	sells, err := s.orm.FindAllSells(model.DirectionSell)
	if err != nil {
		return 0, err
	}

	matchCount, records := s.matchAndComputePnl(sells)
	if err := s.orm.RebuildInventory(); err != nil {
		return matchCount, err
	}
	if err := s.orm.ReplaceAllPnl(records); err != nil {
		return matchCount, err
	}
	return matchCount, nil
}

type pnlKey struct {
	AccountID uint
	Date      string
}

type pnlAgg struct {
	TradeCount int64
	GrossPl    int64
	Fee        int64
	NetPl      int64
}

func (s *service) matchAndComputePnl(sells []model.TradeRecord) (int, []model.PnlDaily) {
	pnlMap := make(map[pnlKey]*pnlAgg)
	matchCount := 0

	for _, sell := range sells {
		remaining := sell.Quantity
		var firstBuyID *uint
		for remaining > 0 {
			buy, err := s.orm.FindEarliestUnmatchedBuy(
				sell.ItemName, sell.Exterior, sell.PaintSeed, sell.PaintIndex, sell.PaintWear, sell.TradeAt,
			)
			if err != nil || buy == nil {
				break
			}

			available := buy.Quantity - buy.ConsumedQuantity
			matchQty := remaining
			if matchQty > available {
				matchQty = available
			}

			if err := s.orm.IncrementConsumedQty(buy.ID, matchQty); err != nil {
				break
			}
			if firstBuyID == nil {
				firstBuyID = &buy.ID
			}

			grossPl := (sell.UnitPrice - buy.UnitPrice) * matchQty
			sellFeeShare := sell.Fee * matchQty / sell.Quantity
			buyFeeShare := buy.Fee * matchQty / buy.Quantity
			fee := sellFeeShare + buyFeeShare
			netPl := grossPl - fee

			date := time.UnixMilli(sell.TradeAt).UTC().Format("2006-01-02")
			key := pnlKey{AccountID: sell.AccountID, Date: date}
			if existing, ok := pnlMap[key]; ok {
				existing.TradeCount++
				existing.GrossPl += grossPl
				existing.Fee += fee
				existing.NetPl += netPl
			} else {
				pnlMap[key] = &pnlAgg{TradeCount: 1, GrossPl: grossPl, Fee: fee, NetPl: netPl}
			}

			remaining -= matchQty
			matchCount++
		}
		if remaining == 0 && firstBuyID != nil {
			_ = s.orm.SetMatchedBuy(sell.ID, *firstBuyID)
		}
	}

	records := make([]model.PnlDaily, 0, len(pnlMap))
	for key, agg := range pnlMap {
		records = append(records, model.PnlDaily{
			AccountID:  key.AccountID,
			Date:       key.Date,
			TradeCount: agg.TradeCount,
			GrossPl:    agg.GrossPl,
			Fee:        agg.Fee,
			NetPl:      agg.NetPl,
		})
	}
	return matchCount, records
}

var Module = fx.Module("pnl",
	logfx.WithComponent("pnl"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) PnlInterface { return s }, fx.As(new(PnlInterface))),
	),
)
