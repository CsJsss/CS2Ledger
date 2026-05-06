package trade

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type CompletedTradeView struct {
	SellTradeID   uint   `json:"sellTradeId"`
	ItemName      string `json:"itemName"`
	SellUnitPrice int64  `json:"sellUnitPrice"`
	Quantity      int64  `json:"quantity"`
	SellFee       int64  `json:"sellFee"`
	SellAt        int64  `json:"sellAt"`
	BuyTradeID    uint   `json:"buyTradeId"`
	BuyUnitPrice  int64  `json:"buyUnitPrice"`
	BuyFee        int64  `json:"buyFee"`
	BuyAt         int64  `json:"buyAt"`
	GrossPl       int64  `json:"grossPl"`
	TotalFee      int64  `json:"totalFee"`
	NetPl         int64  `json:"netPl"`
}

type CompletedTradesSummary struct {
	TotalTrades  int64 `json:"totalTrades"`
	TotalGrossPl int64 `json:"totalGrossPl"`
	TotalFee     int64 `json:"totalFee"`
	TotalNetPl   int64 `json:"totalNetPl"`
}

type TradeInterface interface {
	ListByAccount(accountID uint, tradeType string) ([]model.TradeRecord, error)
	ListCompletedTrades(accountID uint) ([]CompletedTradeView, error)
	GetCompletedTradesSummary(accountID uint) (*CompletedTradesSummary, error)
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
		buys, _ := svc.orm.FindTradesByAccount(accountID, "buy", 0)
		var buy *model.TradeRecord
		for i := range buys {
			if buys[i].ID == *sell.MatchedBuyTradeID {
				buy = &buys[i]
				break
			}
		}
		if buy == nil {
			continue
		}

		grossPl := (sell.UnitPrice - buy.UnitPrice) * sell.Quantity
		views = append(views, CompletedTradeView{
			SellTradeID:   sell.ID,
			ItemName:      sell.ItemName,
			SellUnitPrice: sell.UnitPrice,
			Quantity:      sell.Quantity,
			SellFee:       sell.Fee,
			SellAt:        sell.TradeAt,
			BuyTradeID:    buy.ID,
			BuyUnitPrice:  buy.UnitPrice,
			BuyFee:        buy.Fee,
			BuyAt:         buy.TradeAt,
			GrossPl:       grossPl,
			TotalFee:      sell.Fee + buy.Fee,
			NetPl:         grossPl - (sell.Fee + buy.Fee),
		})
	}
	return views, nil
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
