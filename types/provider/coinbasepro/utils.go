package coinbasepro

import (
	coinbasepro "github.com/preichenberger/go-coinbasepro"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
)

func getStatus(ord coinbasepro.Order) types.OrderStatus {
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

func getType(ord coinbasepro.Order) types.OrderType {
	switch ord.Type {
	case "limit":
		return order.Limit
	case "market":
	}
	return order.Market
}

func getSide(ord coinbasepro.Order) types.OrderSide {
	switch ord.Type {
	case "buy":
		return order.Buy
	case "sell":
	}
	return order.Sell
}
