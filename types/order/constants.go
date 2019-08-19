package order

import "github.com/sinisterminister/currencytrader/types"

const (
	// Pending is for orders still waiting for a watch
	Pending types.OrderStatus = iota

	// Partial is for orders that have been partially filled
	Partial

	// Canceled is for orders that have been cancelled
	Canceled

	// Filled is for orders that have completely filled
	Filled
)

const (
	// Buy represents a buy sided order
	Buy types.OrderSide = iota

	// Sell represents a sell sided order
	Sell
)
