package order

import "github.com/sinisterminister/currencytrader/types"

const (
	// Pending is for orders still waiting for a watch
	Pending types.OrderStatus = "PENDING"

	// Partial is for orders that have been partially filled
	Partial types.OrderStatus = "PARTIAL"

	// Canceled is for orders that have been cancelled
	Canceled types.OrderStatus = "CANCELED"

	// Filled is for orders that have completely filled
	Filled types.OrderStatus = "FILLED"
)

const (
	// Buy represents a buy sided order
	Buy types.OrderSide = "BUY"

	// Sell represents a sell sided order
	Sell types.OrderSide = "SELL"
)
