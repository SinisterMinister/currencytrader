package wallet

import (
	"sync"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/moneytrader/types/currency"
)

// Wallet TODO
type Wallet struct {
	currency currency.Currency
	total    decimal.Decimal
	free     decimal.Decimal
	locked   decimal.Decimal
	reserved decimal.Decimal
	mutex    sync.Mutex
}

func NewWallet(cur currency.Currency, total decimal.Decimal, free decimal.Decimal,
	locked decimal.Decimal, reserved decimal.Decimal) *Wallet {

	return &Wallet{
		currency: cur,
		total:    total,
		free:     free,
		locked:   locked,
		reserved: reserved,
		mutex:    sync.Mutex{},
	}
}

// GetCurrency TODO
func (w *Wallet) GetCurrency() currency.Currency {
	return w.currency
}

// GetTotalBalance TODO
func (w *Wallet) GetTotalBalance() decimal.Decimal {
	return w.total
}

// GetFreeBalance TODO
func (w *Wallet) GetFreeBalance() decimal.Decimal {
	return w.free
}

// GetLockedBalance TODO
func (w *Wallet) GetLockedBalance() decimal.Decimal {
	return w.locked
}

// GetReservedBalance TODO
func (w *Wallet) GetReservedBalance() decimal.Decimal {
	return w.reserved
}
