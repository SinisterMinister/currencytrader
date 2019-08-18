package order

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

// Request represents an order to be placed by the provider
type request struct {
	side     types.OrderSide
	quantity decimal.Decimal
	price    decimal.Decimal
}

func NewRequestFromDTO(dto types.OrderRequestDTO) types.OrderRequest {
	return &request{
		side:     dto.Side,
		quantity: dto.Quantity,
		price:    dto.Price,
	}
}

func NewRequest(side types.OrderSide, quantity decimal.Decimal, price decimal.Decimal) types.OrderRequest {
	return &request{
		side:     side,
		quantity: quantity,
		price:    price,
	}
}

func (req *request) ToDTO() types.OrderRequestDTO {
	return types.OrderRequestDTO{
		Side:     req.Side(),
		Price:    req.Price(),
		Quantity: req.Quantity(),
	}
}

func (r *request) Side() types.OrderSide {
	return r.side
}

func (r *request) Quantity() decimal.Decimal {
	return r.quantity
}

func (r *request) Price() decimal.Decimal {
	return r.price
}
