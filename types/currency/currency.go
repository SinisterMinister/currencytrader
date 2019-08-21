package currency

import "github.com/sinisterminister/currencytrader/types"

// Currency TODO
type currency struct {
	dto types.CurrencyDTO
}

func New(cur types.CurrencyDTO) types.Currency {
	return &currency{cur}
}

func (c *currency) Name() string { return c.dto.Name }

func (c *currency) Symbol() string { return c.dto.Symbol }

func (c *currency) Precision() int { return c.dto.Precision }

func (c *currency) ToDTO() types.CurrencyDTO { return c.dto }
