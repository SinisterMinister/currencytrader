package order

import (
	"time"

	"github.com/shopspring/decimal"
)

// Status handles the various statuses the Order can be in
type Status int

const (
	// Pending is for orders still working to be fulfilled
	Pending Status = iota

	// Canceled is for orders that have been cancelled
	Canceled

	// Success is for orders that have succefully filled
	Success
)

// Order represents an order that has been accepted by the provider
type Order struct {
	CreationTime time.Time
	Filled       decimal.Decimal
	ID           string
	Request      Request
	Status       Status
}
