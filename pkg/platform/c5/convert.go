package c5

import (
	"fmt"
	"strconv"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

func toBuyerTrade(item c5BuyerOrder) platform.TradeRecord {
	price := int64(item.Price * 100)
	fee := int64(item.BuyerFee * 100)
	tradeOfferID := ""
	if item.TradeOfferID != nil {
		tradeOfferID = *item.TradeOfferID
	}
	state := "SUCCESS"
	return platform.TradeRecord{
		ExternalID:   fmt.Sprintf("c5-buy-%s", item.OrderID),
		TradeType:    "buy",
		Quantity:     1,
		UnitPrice:    price,
		TotalPrice:   price,
		Fee:          fee,
		TradeAt:      item.CreateTime * 1000,
		TradeOfferID: tradeOfferID,
		State:        state,
		StateText:    item.StatusName,
	}
}

func toBuyerTradeEnriched(item c5BuyerOrder, detail c5BuyerOrderDetail) platform.TradeRecord {
	trade := toBuyerTrade(item)
	if detail.OfferInfoDTO != nil {
		trade.TradeOfferID = detail.OfferInfoDTO.TradeOfferID
	}
	trade.TransactTime = detail.CreateTime * 1000
	if detail.OpenItemInfo != nil {
		name, exterior := platform.NormalizeItemName(detail.OpenItemInfo.Name)
		trade.CS2Item = model.CS2Item{
			AssetID:  detail.OpenItemInfo.ItemID,
			ItemName: name,
			Exterior: exterior,
		}
	} else {
		trade.CS2Item = model.CS2Item{
			AssetID: detail.NewAssetID,
		}
	}
	if detail.AssetInfo != nil {
		if detail.AssetInfo.AssetID != "" {
			trade.AssetID = detail.AssetInfo.AssetID
		}
		if detail.AssetInfo.InstanceID != "" {
			trade.InstanceID = detail.AssetInfo.InstanceID
		}
		if w, err := strconv.ParseFloat(detail.AssetInfo.Wear, 64); err == nil {
			trade.PaintWear = w
		}
		trade.PaintSeed = detail.AssetInfo.PaintSeed
		trade.PaintIndex = detail.AssetInfo.PaintIndex
	}
	return trade
}

func toSellerTrade(item c5SellerOrder) platform.TradeRecord {
	price := int64(item.Price * 100)
	tradeAt := int64(0)
	statusText := ""
	if item.OrderConfirmInfo != nil {
		tradeAt = item.OrderConfirmInfo.OrderCreateTime * 1000
		statusText = item.OrderConfirmInfo.StatusName
	}
	state := "SUCCESS"
	name, exterior := platform.NormalizeItemName(item.Name)
	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("c5-sell-%s", item.OrderID),
		CS2Item:    model.CS2Item{AssetID: item.ProductID, ItemName: name, Exterior: exterior},
		TradeType:  "sell",
		Quantity:   1,
		UnitPrice:  price,
		TotalPrice: price,
		TradeAt:    tradeAt,
		State:      state,
		StateText:  statusText,
	}
}
