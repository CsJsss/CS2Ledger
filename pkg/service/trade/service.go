package trade

import (
	"errors"
	"sort"

	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

var (
	ErrInvalidSortBy = errors.New("invalid sortBy")
)

// sqlSortBy maps frontend sort keys to SQL column names for ORDER BY.
// These are validated before being passed to the ORM layer.
var sqlSortBy = map[string]string{
	"itemName": "item_name",
	"tradeAt":  "trade_at",
}

// appSortFields are sort keys applied in Go after groups are built.
// These fields are computed per group (sums, counts) and have no SQL column.
var appSortFields = map[string]bool{
	"count":     true,
	"totalBuy":  true,
	"totalSell": true,
	"grossPl":   true,
	"netPl":     true,
	"fees":      true,
}

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

type TradeGroup struct {
	ItemName       string               `json:"itemName"`
	Count          int                  `json:"count"`
	TotalBuyPrice  int64                `json:"totalBuyPrice"`
	TotalSellPrice int64                `json:"totalSellPrice"`
	TotalGrossPl   int64                `json:"totalGrossPl"`
	TotalFee       int64                `json:"totalFee"`
	TotalNetPl     int64                `json:"totalNetPl"`
	Trades         []CompletedTradeView `json:"trades"`
}

type PaginatedGroups struct {
	Groups   []TradeGroup `json:"groups"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
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
	ListCompletedTradeGroups(accountID uint, page, pageSize int, sortBy, sortDir string) (*PaginatedGroups, error)
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

func (svc *service) buildCompletedTradeViews(sells []model.TradeRecord) ([]CompletedTradeView, error) {
	buyIDs := make(map[uint]bool)
	for _, sell := range sells {
		if sell.MatchedBuyTradeID != nil {
			buyIDs[*sell.MatchedBuyTradeID] = true
		}
	}
	ids := make([]uint, 0, len(buyIDs))
	for id := range buyIDs {
		ids = append(ids, id)
	}

	buys, err := svc.orm.FindTradeRecordsByIDs(ids)
	if err != nil {
		return nil, err
	}
	buyMap := make(map[uint]*model.TradeRecord)
	for i := range buys {
		buyMap[buys[i].ID] = &buys[i]
	}

	views := make([]CompletedTradeView, 0, len(sells))
	for _, sell := range sells {
		if sell.MatchedBuyTradeID == nil {
			continue
		}
		buy, ok := buyMap[*sell.MatchedBuyTradeID]
		if !ok {
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

func (svc *service) ListCompletedTradeGroups(accountID uint, page, pageSize int, sortBy, sortDir string) (*PaginatedGroups, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "asc"
	}

	// SQL-level sort: paginate group names via ORM, keep only current page
	if col, ok := sqlSortBy[sortBy]; ok {
		offset := (page - 1) * pageSize
		names, total, err := svc.orm.FindCompletedTradeGroupNames(accountID, offset, pageSize, col, sortDir)
		if err != nil {
			return nil, err
		}
		groups, err := svc.fetchAndBuildGroups(accountID, names)
		if err != nil {
			return nil, err
		}
		return &PaginatedGroups{Groups: groups, Total: total, Page: page, PageSize: pageSize}, nil
	}

	// App-level sort: fetch all groups, sort in Go, then paginate
	if appSortFields[sortBy] {
		names, _, err := svc.orm.FindCompletedTradeGroupNames(accountID, 0, 10000, "item_name", "asc")
		if err != nil {
			return nil, err
		}
		allGroups, err := svc.fetchAndBuildGroups(accountID, names)
		if err != nil {
			return nil, err
		}
		svc.sortGroups(allGroups, sortBy, sortDir)
		total := int64(len(allGroups))
		offset := (page - 1) * pageSize
		if offset >= len(allGroups) {
			return &PaginatedGroups{Groups: nil, Total: total, Page: page, PageSize: pageSize}, nil
		}
		end := offset + pageSize
		if end > len(allGroups) {
			end = len(allGroups)
		}
		return &PaginatedGroups{Groups: allGroups[offset:end], Total: total, Page: page, PageSize: pageSize}, nil
	}

	return nil, ErrInvalidSortBy
}

func (svc *service) fetchAndBuildGroups(accountID uint, names []string) ([]TradeGroup, error) {
	if len(names) == 0 {
		return nil, nil
	}
	sells, err := svc.orm.FindSellsByItemNames(accountID, names)
	if err != nil {
		return nil, err
	}
	views, err := svc.buildCompletedTradeViews(sells)
	if err != nil {
		return nil, err
	}

	byName := make(map[string][]CompletedTradeView)
	for _, v := range views {
		n := v.ItemName
		byName[n] = append(byName[n], v)
	}

	groups := make([]TradeGroup, 0, len(names))
	for _, name := range names {
		trades := byName[name]
		var totalBuyPrice, totalSellPrice, totalGrossPl, totalFee, totalNetPl int64
		for _, t := range trades {
			totalBuyPrice += t.BuyTrade.TotalPrice
			totalSellPrice += t.SellTrade.TotalPrice
			totalGrossPl += t.GrossPl
			totalFee += t.TotalFee
			totalNetPl += t.NetPl
		}
		groups = append(groups, TradeGroup{
			ItemName:       name,
			Count:          len(trades),
			TotalBuyPrice:  totalBuyPrice,
			TotalSellPrice: totalSellPrice,
			TotalGrossPl:   totalGrossPl,
			TotalFee:       totalFee,
			TotalNetPl:     totalNetPl,
			Trades:         trades,
		})
	}
	return groups, nil
}

func (svc *service) sortGroups(groups []TradeGroup, sortBy, sortDir string) {
	desc := sortDir == "desc"
	sort.Slice(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		var less bool
		switch sortBy {
		case "count":
			less = a.Count < b.Count
		case "totalBuy":
			less = a.TotalBuyPrice < b.TotalBuyPrice
		case "totalSell":
			less = a.TotalSellPrice < b.TotalSellPrice
		case "grossPl":
			less = a.TotalGrossPl < b.TotalGrossPl
		case "netPl":
			less = a.TotalNetPl < b.TotalNetPl
		case "fees":
			less = a.TotalFee < b.TotalFee
		default:
			less = a.ItemName < b.ItemName
		}
		if desc {
			return !less
		}
		return less
	})
}

func (svc *service) ListCompletedTrades(accountID uint) ([]CompletedTradeView, error) {
	sells, err := svc.orm.FindSellsWithMatchedBuy(accountID)
	if err != nil {
		return nil, err
	}
	return svc.buildCompletedTradeViews(sells)
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
