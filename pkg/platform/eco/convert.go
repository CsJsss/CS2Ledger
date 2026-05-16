package eco

import (
	"fmt"
	"strconv"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

var cst = time.FixedZone("CST", 8*3600)

var tradeAtFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02",
}

func yuanToFen(yuan float64) int64 {
	return int64(yuan * 100)
}

func parseTradeAt(s string) int64 {
	for _, f := range tradeAtFormats {
		t, err := time.ParseInLocation(f, s, cst)
		if err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func toBuyTradeFromListItem(o buyerOrderModel) platform.TradeRecord {
	name, exterior := platform.NormalizeItemName(o.GoodsName)
	unitPrice := o.OrderAmount
	if o.GoodsTotal > 1 {
		unitPrice = o.OrderAmount / float64(o.GoodsTotal)
	}
	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("eco-buy-%s", o.OrderNum),
		CS2Item: model.CS2Item{
			MarketHashName: o.HashName,
			ItemName:       name,
			Exterior:       exterior,
		},
		TradeType:  model.DirectionBuy,
		Quantity:   int64(o.GoodsTotal),
		UnitPrice:  yuanToFen(unitPrice),
		TotalPrice: yuanToFen(o.OrderAmount),
		TradeAt:    parseTradeAt(o.CreateOrderTime),
	}
}

func toSellTradeFromListItem(o sellerOrderModel) platform.TradeRecord {
	name, exterior := platform.NormalizeItemName(o.GoodsName)
	unitPrice := o.OrderAmount
	if o.GoodsTotal > 1 {
		unitPrice = o.OrderAmount / float64(o.GoodsTotal)
	}
	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("eco-sell-%s", o.OrderNum),
		CS2Item: model.CS2Item{
			MarketHashName: o.HashName,
			ItemName:       name,
			Exterior:       exterior,
		},
		TradeType:  model.DirectionSell,
		Quantity:   int64(o.GoodsTotal),
		UnitPrice:  yuanToFen(unitPrice),
		TotalPrice: yuanToFen(o.OrderAmount),
		TradeAt:    parseTradeAt(o.CreateOrderTime),
	}
}

func toCS2Item(ap assetPreviewModel) model.CS2Item {
	name, exterior := platform.NormalizeItemName(ap.GoodsName)
	pw, _ := strconv.ParseFloat(ap.PaintWear, 64)
	stickers := make([]model.Sticker, 0, len(ap.Stickers))
	for _, s := range ap.Stickers {
		stickers = append(stickers, model.Sticker{
			StickerID:      s.Id,
			Slot:           s.Slot,
			Wear:           s.Wear,
			Name:           s.Name,
			ImageURL:       s.Icon,
			ReferencePrice: s.Price,
		})
	}
	keychains := make([]model.Keychain, 0, len(ap.Keychains))
	for _, k := range ap.Keychains {
		keychains = append(keychains, model.Keychain{
			Name:     k.Name,
			ImageURL: k.Icon,
		})
	}
	return model.CS2Item{
		AssetID:        ap.AssetId,
		MarketHashName: ap.HashName,
		ItemName:       name,
		IconURL:        ap.GoodsImage,
		PaintWear:      pw,
		PaintSeed:      ap.PaintSeed,
		PaintIndex:     ap.PaintIndex,
		Exterior:       exterior,
		Stickers:       stickers,
		Keychains:      keychains,
	}
}
