package youpin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

// parseYoupinPrice converts a youpin price value to 分 (cents).
// Before a certain date the API returned 元 (e.g. "1610.00"), after that it returns 分 (e.g. "161000").
// Values with a decimal point are in 元 and need ×100.
func parseYoupinPrice(n json.Number) int64 {
	s := n.String()
	if strings.Contains(s, ".") {
		v, err := n.Float64()
		if err != nil {
			return 0
		}
		return int64(v * 100)
	}
	v, err := n.Int64()
	if err != nil {
		return 0
	}
	return v
}

func toBuyTrade(o youpinBuyOrder, p youpinBuyProduct) platform.TradeRecord {
	assetID := fmt.Sprintf("%d", p.AssertID)
	if p.AssertID == 0 {
		assetID = fmt.Sprintf("%d", p.CommodityID)
	}
	qty := int64(o.CommodityNum)
	if qty == 0 {
		qty = 1
	}
	totalPrice := o.TotalAmount
	unitPrice := totalPrice / qty
	name, ext := platform.NormalizeItemName(p.CommodityName)
	exterior := p.ExteriorName
	if exterior == "" {
		exterior = ext
	}
	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("youpin-buy-%d-%d", o.OrderID, p.AssertID),
		CS2Item: model.CS2Item{
			AssetID: assetID, ItemName: name, MarketHashName: p.CommodityHashName,
			Exterior: exterior, PaintWear: parseWear(p.CommodityAbrade),
			Rarity: p.RarityName, WeaponType: p.TypeName, Itemset: p.ItemSetName,
			PaintSeed: p.PaintSeed, PaintIndex: p.PaintIndex,
		},
		TradeType:  model.DirectionBuy,
		Quantity:   qty,
		UnitPrice:  unitPrice,
		TotalPrice: totalPrice,
		Fee:        0,
		TradeAt:    o.FinishOrderTime,
	}
}

func toSellTrade(o youpinSellOrder) platform.TradeRecord {
	qty := int64(o.CommodityNum)
	if qty == 0 {
		qty = 1
	}
	paymentAmount := parseYoupinPrice(o.PaymentAmount)
	fee := parseYoupinPrice(o.TotalFeeAmount)
	unitPrice := paymentAmount / qty
	name, ext := platform.NormalizeItemName(o.ProductDetail.CommodityName)
	exterior := o.ProductDetail.ExteriorName
	if exterior == "" {
		exterior = ext
	}
	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("youpin-sell-%s", o.OrderNo),
		CS2Item: model.CS2Item{
			AssetID:        fmt.Sprintf("%d", o.ProductDetail.AssertID),
			ItemName:       name,
			MarketHashName: o.ProductDetail.CommodityHashName,
			Exterior:       exterior,
			PaintWear:      parseWear(o.ProductDetail.CommodityAbrade),
			Rarity:         o.ProductDetail.RarityName,
			WeaponType:     o.ProductDetail.TypeName,
			Itemset:        o.ProductDetail.ItemSetName,
			PaintSeed:      o.ProductDetail.PaintSeed,
			PaintIndex:     o.ProductDetail.PaintIndex,
		},
		TradeType:  model.DirectionSell,
		Quantity:   qty,
		UnitPrice:  unitPrice,
		TotalPrice: paymentAmount,
		Fee:        fee,
		TradeAt:    o.FinishOrderTime,
	}
}

// youpinTypeToInternal maps YouPin typeId to internal bill type constants.
func youpinTypeToInternal(typeID int) int {
	switch typeID {
	case 3:
		return model.BillTypePurchase
	case 16:
		return model.BillTypeRentalIncome
	case 187:
		return model.BillTypeRentalFee
	default:
		return model.BillTypeOther
	}
}

func toBillRecord(item youpinBillItem) (platform.BillRecord, error) {
	money := parseYoupinPrice(json.Number(item.ThisMoney))
	t, err := time.ParseInLocation("2006-01-02 15:04:05", item.AddTime, time.Local)
	if err != nil {
		return platform.BillRecord{}, fmt.Errorf("parse addTime %q: %w", item.AddTime, err)
	}
	addTimeMs := t.UnixMilli()

	return platform.BillRecord{
		TypeName:  item.TypeName,
		TypeID:    youpinTypeToInternal(item.TypeID),
		ThisMoney: money,
		OrderNo:   item.OrderNo,
		AddTime:   addTimeMs,
	}, nil
}
