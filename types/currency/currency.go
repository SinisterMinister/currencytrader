package currency

import "github.com/sinisterminister/currencytrader/types"

// Currency TODO
type currency struct {
	name      string
	symbol    string
	precision int
}

func (c *currency) Name() string {
	return c.name
}

func (c *currency) Symbol() string {
	return c.symbol
}

func (c *currency) Precision() int {
	return c.precision
}

func New(cur types.CurrencyDTO) types.Currency {
	return &currency{
		name:      cur.Name,
		symbol:    cur.Symbol,
		precision: cur.Precision,
	}
}

func ToDTO(cur types.Currency) types.CurrencyDTO {
	return types.CurrencyDTO{
		Name:      cur.Name(),
		Symbol:    cur.Symbol(),
		Precision: cur.Precision(),
	}
}
