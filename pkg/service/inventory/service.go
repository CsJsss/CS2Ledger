package inventory

import (
	"errors"
	"sort"

	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/utils/dateutil"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

var (
	ErrInvalidSortBy = errors.New("invalid sortBy")
)

func ptrValue(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// sqlSortBy maps frontend sort keys to SQL column names for ORDER BY.
var sqlSortBy = map[string]string{
	"itemName":  "item_name",
	"updatedAt": "updated_at",
}

// appSortFields are sort keys applied in Go after groups are built.
var appSortFields = map[string]bool{
	"count":         true,
	"totalQuantity": true,
	"totalBuyPrice": true,
	"avgBuyPrice":   true,
	"marketPrice":   true,
	"marketTotal":   true,
	"unrealizedPl":  true,
	"plPercent":     true,
}

// InventoryGroup represents a grouped set of inventory items by item name + exterior.
type InventoryGroup struct {
	ItemName             string                `json:"itemName"`
	Exterior             string                `json:"exterior"`
	CsqaqGoodsID         int                   `json:"csqaqGoodsId,omitempty"`
	MarketHashName       string                `json:"marketHashName"`
	WeaponType           string                `json:"weaponType"`
	Count                int                   `json:"count"`
	TotalQuantity        int64                 `json:"totalQuantity"`
	TotalBuyPrice        int64                 `json:"totalBuyPrice"`
	AvgBuyPrice          int64                 `json:"avgBuyPrice"`
	MarketPrice          *int64                `json:"marketPrice,omitempty"`
	MarketPriceUpdatedAt *int64                `json:"marketPriceUpdatedAt,omitempty"`
	UnrealizedPl         *int64                `json:"unrealizedPl,omitempty"`
	Instances            []model.InventoryItem `json:"instances"`
}

type PaginatedGroups struct {
	Groups   []InventoryGroup `json:"groups"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type InventoryInterface interface {
	List(accountID uint, status string) ([]model.InventoryItem, error)
	GetItemDetail(accountID uint, assetID string) (*ItemDetail, error)
	ListGroups(accountID uint, status, weaponType string, page, pageSize int, sortBy, sortDir string) (*PaginatedGroups, error)
	ListDailyBuys(accountID uint) ([]DailyBuyGroup, error)
	SetPriceProvider(p PriceProvider)
	SetPriceSource(source string)
}

// PriceProvider is a subset of market.MarketInterface to avoid import cycles.
type PriceProvider interface {
	GetAllPrices() ([]platform.PriceInfo, error)
}

type service struct {
	log         *logfx.Logger
	orm         orm.ORMInterface
	prices      PriceProvider
	priceSource string
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

// SetPriceProvider sets the market price source (called after DI wiring).
func (s *service) SetPriceProvider(p PriceProvider) {
	s.prices = p
}

// SetPriceSource sets which price to use (buff/youpin/steam).
func (s *service) SetPriceSource(source string) {
	s.priceSource = source
}

// resolvePriceMap fetches market prices and returns a map from MarketHashName to price in fen.
// Returns nil if prices are unavailable (best-effort).
func (s *service) resolvePriceMap() map[string]int64 {
	if s.prices == nil {
		return nil
	}
	priceList, err := s.prices.GetAllPrices()
	if err != nil {
		return nil
	}
	priceMap := make(map[string]int64, len(priceList))
	for _, p := range priceList {
		var mp int64
		switch s.priceSource {
		case "youpin":
			mp = int64(p.YoupinPrice * 100)
		case "steam":
			mp = int64(p.SteamPrice * 100)
		default:
			mp = int64(p.BuffPrice * 100)
		}
		priceMap[p.MarketHashName] = mp
	}
	return priceMap
}

func (s *service) List(accountID uint, status string) ([]model.InventoryItem, error) {
	return s.orm.FindInventoryByAccount(accountID, status)
}

func (s *service) ListGroups(accountID uint, status, weaponType string, page, pageSize int, sortBy, sortDir string) (*PaginatedGroups, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "asc"
	}

	// SQL-level sort: paginate group keys via ORM, keep only current page
	if col, ok := sqlSortBy[sortBy]; ok {
		offset := (page - 1) * pageSize
		keys, total, err := s.orm.FindInventoryGroupKeys(accountID, status, weaponType, offset, pageSize, col, sortDir)
		if err != nil {
			return nil, err
		}
		groups, err := s.buildGroups(accountID, keys)
		if err != nil {
			return nil, err
		}
		s.enrichWithPrices(groups)
		return &PaginatedGroups{Groups: groups, Total: total, Page: page, PageSize: pageSize}, nil
	}

	// App-level sort: fetch all groups, sort in Go, then paginate
	if appSortFields[sortBy] {
		keys, _, err := s.orm.FindInventoryGroupKeys(accountID, status, weaponType, 0, 10000, "item_name", "asc")
		if err != nil {
			return nil, err
		}
		allGroups, err := s.buildGroups(accountID, keys)
		if err != nil {
			return nil, err
		}
		s.enrichWithPrices(allGroups)
		s.sortGroups(allGroups, sortBy, sortDir)
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

func (s *service) enrichWithPrices(groups []InventoryGroup) {
	if s.prices == nil || len(groups) == 0 {
		return
	}
	priceList, err := s.prices.GetAllPrices()
	if err != nil || len(priceList) == 0 {
		return
	}
	priceMap := make(map[string]platform.PriceInfo, len(priceList))
	for _, p := range priceList {
		priceMap[p.MarketHashName] = p
	}
	for i := range groups {
		g := &groups[i]
		if g.MarketHashName == "" || g.AvgBuyPrice == 0 {
			continue
		}
		info, ok := priceMap[g.MarketHashName]
		if !ok {
			continue
		}
		var mp int64
		switch s.priceSource {
		case "youpin":
			mp = int64(info.YoupinPrice * 100)
		case "steam":
			mp = int64(info.SteamPrice * 100)
		default:
			mp = int64(info.BuffPrice * 100)
		}
		g.MarketPrice = &mp
		updatedAt := info.UpdatedAt
		g.MarketPriceUpdatedAt = &updatedAt
		upl := (mp - g.AvgBuyPrice) * g.TotalQuantity
		g.UnrealizedPl = &upl
	}
}

func (s *service) buildGroups(accountID uint, keys []orm.InventoryGroupKey) ([]InventoryGroup, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	instances, err := s.orm.FindInventoryByGroupKeys(accountID, keys)
	if err != nil {
		return nil, err
	}

	type groupKey struct {
		itemName string
		exterior string
	}
	byKey := make(map[groupKey][]model.InventoryItem)
	for _, inst := range instances {
		k := groupKey{itemName: inst.ItemName, exterior: inst.Exterior}
		byKey[k] = append(byKey[k], inst)
	}

	groups := make([]InventoryGroup, 0, len(keys))
	for _, k := range keys {
		gk := groupKey{itemName: k.ItemName, exterior: k.Exterior}
		insts := byKey[gk]
		wt := ""
		mhn := ""
		goodsID := 0
		var totalQty int64
		var totalBuyPrice int64
		for _, inst := range insts {
			qty := inst.Quantity
			if qty == 0 {
				qty = 1
			}
			totalQty += qty
			if inst.BuyTrade != nil {
				totalBuyPrice += inst.BuyTrade.UnitPrice * qty
			}
			if mhn == "" {
				mhn = inst.MarketHashName
			}
			if goodsID == 0 {
				goodsID = inst.CsqaqGoodsID
			}
		}
		var avgBuyPrice int64
		if totalQty > 0 {
			avgBuyPrice = totalBuyPrice / totalQty
		}
		if len(insts) > 0 {
			wt = insts[0].WeaponType
		}
		groups = append(groups, InventoryGroup{
			ItemName:       k.ItemName,
			Exterior:       k.Exterior,
			CsqaqGoodsID:   goodsID,
			MarketHashName: mhn,
			WeaponType:     wt,
			Count:          len(insts),
			TotalQuantity:  totalQty,
			TotalBuyPrice:  totalBuyPrice,
			AvgBuyPrice:    avgBuyPrice,
			Instances:      insts,
		})
	}
	return groups, nil
}

func (s *service) sortGroups(groups []InventoryGroup, sortBy, sortDir string) {
	desc := sortDir == "desc"
	sort.Slice(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		var less bool
		switch sortBy {
		case "count":
			less = a.Count < b.Count
		case "totalBuyPrice":
			less = a.TotalBuyPrice < b.TotalBuyPrice
		case "avgBuyPrice":
			less = a.AvgBuyPrice < b.AvgBuyPrice
		case "totalQuantity":
			less = a.TotalQuantity < b.TotalQuantity
		case "marketPrice":
			less = ptrValue(a.MarketPrice) < ptrValue(b.MarketPrice)
		case "marketTotal":
			less = marketTotal(a) < marketTotal(b)
		case "unrealizedPl":
			less = ptrValue(a.UnrealizedPl) < ptrValue(b.UnrealizedPl)
		case "plPercent":
			less = plPercent(a) < plPercent(b)
		default:
			less = a.ItemName < b.ItemName
		}
		if desc {
			return !less
		}
		return less
	})
}

func (s *service) ListDailyBuys(accountID uint) ([]DailyBuyGroup, error) {
	rows, err := s.orm.FindDailyBuys(accountID)
	if err != nil {
		return nil, err
	}

	priceMap := s.resolvePriceMap()

	type dateKey string
	byDate := make(map[dateKey][]DailyBuyItem)
	for _, r := range rows {
		date, _ := dateutil.FormatTimestamp(r.BuyAt)
		dk := dateKey(date)
		totalCost := r.BuyPrice * r.Quantity
		item := DailyBuyItem{
			ItemName:  r.ItemName,
			Exterior:  r.Exterior,
			Quantity:  r.Quantity,
			BuyPrice:  r.BuyPrice,
			TotalCost: totalCost,
			Platform:  r.Source,
			Status:    r.Status,
		}
		if priceMap != nil {
			if mp, ok := priceMap[r.MarketHashName]; ok {
				item.MarketPrice = &mp
				upl := (mp - r.BuyPrice) * r.Quantity
				item.UnrealizedPl = &upl
			}
		}
		byDate[dk] = append(byDate[dk], item)
	}

	groups := make([]DailyBuyGroup, 0, len(byDate))
	for dk, items := range byDate {
		var totalCost int64
		var totalMV int64
		hasMV := true
		for _, it := range items {
			totalCost += it.TotalCost
			if it.MarketPrice != nil {
				totalMV += *it.MarketPrice * it.Quantity
			} else {
				hasMV = false
			}
		}
		t, _ := dateutil.ParseDate(string(dk))
		g := DailyBuyGroup{
			Date:       string(dk),
			DayOfWeek:  dateutil.DayOfWeekNames[t.Weekday()],
			Items:      items,
			TotalCount: len(items),
			TotalCost:  totalCost,
		}
		if hasMV && len(items) > 0 {
			g.TotalMarketValue = &totalMV
		}
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Date > groups[j].Date })
	return groups, nil
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

type DailyBuyItem struct {
	ItemName     string `json:"itemName"`
	Exterior     string `json:"exterior"`
	Quantity     int64  `json:"quantity"`
	BuyPrice     int64  `json:"buyPrice"`
	TotalCost    int64  `json:"totalCost"`
	MarketPrice  *int64 `json:"marketPrice,omitempty"`
	UnrealizedPl *int64 `json:"unrealizedPl,omitempty"`
	Platform     string `json:"platform"`
	Status       string `json:"status"`
}

type DailyBuyGroup struct {
	Date             string         `json:"date"`
	DayOfWeek        string         `json:"dayOfWeek"`
	Items            []DailyBuyItem `json:"items"`
	TotalCount       int            `json:"totalCount"`
	TotalCost        int64          `json:"totalCost"`
	TotalMarketValue *int64         `json:"totalMarketValue,omitempty"`
}

var Module = fx.Module("inventory",
	logfx.WithComponent("inventory"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) InventoryInterface { return s }, fx.As(new(InventoryInterface))),
	),
)

func plPercent(g InventoryGroup) float64 {
	if g.MarketPrice == nil || g.AvgBuyPrice == 0 {
		return -1e18
	}
	return float64(*g.MarketPrice-g.AvgBuyPrice) / float64(g.AvgBuyPrice)
}

func marketTotal(g InventoryGroup) int64 {
	return ptrValue(g.MarketPrice) * g.TotalQuantity
}
