package order

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

// Request represents an order to be placed by the provider
type request struct {
	dto    types.OrderRequestDTO
	market types.Market
}

func NewRequestFromDTO(m types.Market, dto types.OrderRequestDTO) types.OrderRequest {
	return &request{
		dto:    dto,
		market: m,
	}
}

func NewRequest(m types.Market, t types.OrderType, s types.OrderSide, quantity decimal.Decimal, price decimal.Decimal, forceMaker bool) types.OrderRequest {
	return &request{
		dto: types.OrderRequestDTO{
			Type:       t,
			Side:       s,
			Price:      price,
			Quantity:   quantity,
			Market:     m.ToDTO(),
			ForceMaker: forceMaker,
		},
		market: m,
	}
}

func (r *request) ToDTO() types.OrderRequestDTO {
	return r.dto
}

func (r *request) Side() types.OrderSide { return r.dto.Side }

func (r *request) Quantity() decimal.Decimal { return r.dto.Quantity }

func (r *request) Price() decimal.Decimal { return r.dto.Price }

func (r *request) Type() types.OrderType { return r.dto.Type }

func (r *request) Market() types.Market { return r.market }

func (r *request) ForceMaker() bool { return r.dto.ForceMaker }
