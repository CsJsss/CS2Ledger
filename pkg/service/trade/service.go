package trade

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type CompletedTradeView struct {
	ItemName  string  `json:"itemName"`
	Exterior  string  `json:"exterior"`
	PaintWear float64 `json:"paintWear"`
	Quantity  int64   `json:"quantity"`
	GrossPl   int64   `json:"grossPl"`
	TotalFee  int64   `json:"totalFee"`
	NetPl     int64   `json:"netPl"`

	SellTrade model.TradeRecord `json:"sellTrade"`
	BuyTrade  model.TradeRecord `json:"buyTrade"`
}

type CompletedTradesSummary struct {
	TotalTrades       int64 `json:"totalTrades"`
	TotalGrossPl      int64 `json:"totalGrossPl"`
	TotalFee          int64 `json:"totalFee"`
	TotalNetPl        int64 `json:"totalNetPl"`
	WithdrawalFee     int64 `json:"withdrawalFee"`
	WithdrawalFeeRate int64 `json:"withdrawalFeeRate"`
}

type TradeInterface interface {
	ListByAccount(accountID uint, tradeType string) ([]model.TradeRecord, error)
	ListCompletedTrades(accountID uint) ([]CompletedTradeView, error)
	GetCompletedTradesSummary(accountID uint) (*CompletedTradesSummary, error)
	ListUnmatchedSells(accountID uint) ([]model.TradeRecord, error)
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (svc *service) ListByAccount(accountID uint, tradeType string) ([]model.TradeRecord, error) {
	return svc.orm.FindTradesByAccount(accountID, tradeType, 0)
}

func (svc *service) ListCompletedTrades(accountID uint) ([]CompletedTradeView, error) {
	sells, err := svc.orm.FindSellsWithMatchedBuy(accountID)
	if err != nil {
		return nil, err
	}

	views := make([]CompletedTradeView, 0, len(sells))
	for _, sell := range sells {
		buy, err := svc.orm.FindMatchedBuyForSell(sell.ID)
		if err != nil || buy == nil {
			continue
		}

		grossPl := (sell.UnitPrice - buy.UnitPrice) * sell.Quantity
		views = append(views, CompletedTradeView{
			ItemName:  sell.ItemName,
			Exterior:  sell.Exterior,
			PaintWear: sell.PaintWear,
			Quantity:  sell.Quantity,
			GrossPl:   grossPl,
			TotalFee:  sell.Fee + buy.Fee,
			NetPl:     grossPl - (sell.Fee + buy.Fee),
			SellTrade: sell,
			BuyTrade:  *buy,
		})
	}
	return views, nil
}

func (svc *service) ListUnmatchedSells(accountID uint) ([]model.TradeRecord, error) {
	return svc.orm.FindUnmatchedSells(accountID)
}

func (svc *service) GetCompletedTradesSummary(accountID uint) (*CompletedTradesSummary, error) {
	views, err := svc.ListCompletedTrades(accountID)
	if err != nil {
		return nil, err
	}
	sum := &CompletedTradesSummary{}
	for _, t := range views {
		sum.TotalTrades++
		sum.TotalGrossPl += t.GrossPl
		sum.TotalFee += t.TotalFee
		sum.TotalNetPl += t.NetPl
	}
	return sum, nil
}

var Module = fx.Module("trade",
	logfx.WithComponent("trade"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) TradeInterface { return s }, fx.As(new(TradeInterface))),
	),
)
