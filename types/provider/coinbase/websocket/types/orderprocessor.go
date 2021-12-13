package types

import "github.com/sinisterminister/currencytrader/types"

type OrderProcessor interface {
	Output() <-chan types.OrderDTO
}
