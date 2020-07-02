package fees

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

// Currency TODO
type fees struct {
	trader internal.Trader
	dto    types.FeesDTO
}

func New(trader internal.Trader, dto types.FeesDTO) types.Fees {
	return &fees{trader, dto}
}

func (f *fees) MakerRate() decimal.Decimal { return f.dto.MakerRate }

func (f *fees) TakerRate() decimal.Decimal { return f.dto.TakerRate }

func (f *fees) Volume() decimal.Decimal { return f.dto.Volume }

func (f *fees) ToDTO() types.FeesDTO { return f.dto }
