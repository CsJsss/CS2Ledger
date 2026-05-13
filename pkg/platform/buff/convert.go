package buff

import (
	"fmt"
	"strconv"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

func toBuyTrade(item buyOrderItem, goodsInfos map[string]goodInfo) platform.TradeRecord {
	goodsIDStr := strconv.FormatInt(item.GoodsID, 10)
	var goods *goodInfo
	if gi, ok := goodsInfos[goodsIDStr]; ok {
		goods = &gi
	}
	price, _ := strconv.ParseFloat(item.Price, 64)
	fee, _ := strconv.ParseFloat(item.Fee, 64)
	return platform.TradeRecord{
		CS2Item:      toCS2Item(item.AssetInfo, item.GoodsID, goods),
		ExternalID:   fmt.Sprintf("buff-buy-%s", item.AssetInfo.AssetID),
		TradeType:    model.DirectionBuy,
		Quantity:     1,
		UnitPrice:    int64(price * 100),
		TotalPrice:   int64(price * 100),
		Fee:          int64(fee * 100),
		TradeAt:      item.TransactTime * 1000,
		State:        item.State,
		StateText:    item.StateText,
		TransactTime: item.TransactTime * 1000,
		TradeOfferID: item.TradeOfferID,
	}
}

func toSellTrade(item sellOrderItem, goodsInfos map[string]goodInfo) platform.TradeRecord {
	goodsIDStr := strconv.FormatInt(item.GoodsID, 10)
	var goods *goodInfo
	if gi, ok := goodsInfos[goodsIDStr]; ok {
		goods = &gi
	}
	price, _ := strconv.ParseFloat(item.Price, 64)
	fee, _ := strconv.ParseFloat(item.Fee, 64)
	income, _ := strconv.ParseFloat(item.Income, 64)
	return platform.TradeRecord{
		CS2Item:      toCS2Item(item.AssetInfo, item.GoodsID, goods),
		ExternalID:   fmt.Sprintf("buff-sell-%s", item.ID),
		TradeType:    model.DirectionSell,
		Quantity:     1,
		UnitPrice:    int64(price * 100),
		TotalPrice:   int64(income * 100),
		Fee:          int64(fee * 100),
		TradeAt:      item.TransactTime * 1000,
		State:        item.State,
		StateText:    item.StateText,
		TransactTime: item.TransactTime * 1000,
		TradeOfferID: item.TradeOfferID,
	}
}
