package coinbase

import (
	cbp "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
)

func getStatus(ord cbp.Order) types.OrderStatus {
	filled, _ := decimal.NewFromString(ord.FilledSize)
	switch ord.Status {
	case "received":
		return order.Pending
	case "open":
		if filled.IsZero() {
			return order.Pending
		}
		return order.Partial
	case "done":
		if ord.DoneReason == "filled" {
			return order.Filled
		}
		return order.Canceled
	}

	return order.Unknown
}

func getType(ord cbp.Order) types.OrderType {
	switch ord.Type {
	case "limit":
		return order.Limit
	case "market":
	}
	return order.Market
}

func getSide(ord cbp.Order) types.OrderSide {
	switch ord.Type {
	case "buy":
		return order.Buy
	case "sell":
	}
	return order.Sell
}
