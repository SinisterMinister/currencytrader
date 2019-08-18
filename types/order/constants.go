package order

import "github.com/sinisterminister/currencytrader/types"

const (
	// Pending is for orders still working to be fulfilled
	Pending types.OrderStatus = iota

	// Canceled is for orders that have been cancelled
	Canceled

	// Success is for orders that have succefully filled
	Success
)

const (
	// BuySide represents a buy sided order
	BuySide types.OrderSide = iota

	// SellSide represents a sell sided order
	SellSide
)
