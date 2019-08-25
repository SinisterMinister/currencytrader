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

	// Rejected is for orders that been rejected
	Rejected types.OrderStatus = "REJECTED"

	// Expired is for order that expired
	Expired types.OrderStatus = "EXPIRED"

	// Unknown is for an order that is of unknown status
	Unknown types.OrderStatus = "UNKNOWN"
)

const (
	// Buy represents a buy sided order
	Buy types.OrderSide = "BUY"

	// Sell represents a sell sided order
	Sell types.OrderSide = "SELL"
)

const (
	// Market represents a market order
	Market types.OrderType = "MARKET"

	// Limit represents a limit order
	Limit types.OrderType = "LIMIT"
)
