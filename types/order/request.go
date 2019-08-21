package order

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

// Request represents an order to be placed by the provider
type request struct {
	dto types.OrderRequestDTO
}

func NewRequestFromDTO(dto types.OrderRequestDTO) types.OrderRequest {
	return &request{dto}
}

func NewRequest(side types.OrderSide, quantity decimal.Decimal, price decimal.Decimal) types.OrderRequest {
	return &request{types.OrderRequestDTO{
		Side:     side,
		Price:    price,
		Quantity: quantity,
	}}
}

func (r *request) ToDTO() types.OrderRequestDTO {
	return r.dto
}

func (r *request) Side() types.OrderSide { return r.dto.Side }

func (r *request) Quantity() decimal.Decimal { return r.dto.Quantity }

func (r *request) Price() decimal.Decimal { return r.dto.Price }
