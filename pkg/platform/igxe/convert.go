package igxe

import (
	"fmt"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

func toSellTrade(item igxeOrder, tradeAt int64) platform.TradeRecord {
	price := int64(item.Price * 100)
	name, exterior := platform.NormalizeItemName(item.GoodsName)
	return platform.TradeRecord{
		ExternalID: fmt.Sprintf("igxe-sell-%s", item.OrderNum),
		CS2Item:    model.CS2Item{AssetID: item.AssetID, ItemName: name, Exterior: exterior},
		TradeType:  model.DirectionSell,
		Quantity:   1,
		UnitPrice:  price,
		TotalPrice: price,
		Fee:        0,
		TradeAt:    tradeAt,
	}
}
