package wallet

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/currency"
	"github.com/sinisterminister/currencytrader/types/internal"
)

// wallet TODO
type wallet struct {
	currency types.Currency

	mutex    sync.RWMutex
	free     decimal.Decimal
	locked   decimal.Decimal
	reserved decimal.Decimal
}

func New(dto types.WalletDTO) internal.Wallet {
	return &wallet{
		currency: currency.New(dto.Currency),
		free:     dto.Free,
		locked:   dto.Locked,
		reserved: dto.Reserved,
	}
}

func (w *wallet) Currency() types.Currency {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.currency
}

func (w *wallet) Total() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.free.Add(w.locked)
}

func (w *wallet) Free() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.free
}

func (w *wallet) Locked() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.locked
}

func (w *wallet) Reserved() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.reserved
}

func (w *wallet) Available() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.free.Sub(w.reserved)
}

func (w *wallet) Release(amount decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.reserved.LessThan(amount) {
		return errors.New("not enough reserved funds to release")
	}

	w.reserved = w.reserved.Sub(amount)
	return nil
}
func (w *wallet) Reserve(amount decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.free.Sub(w.reserved).LessThan(amount) {
		return errors.New("not enough available funds to freeze")
	}

	w.reserved = w.reserved.Add(amount)
	return nil
}

func (w *wallet) UpdateWallet(dto types.WalletDTO) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Skip reserved
	w.free = dto.Free
	w.locked = dto.Locked
}
