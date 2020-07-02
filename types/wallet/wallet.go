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
	trader internal.Trader

	mutex sync.RWMutex
	dto   types.WalletDTO
}

func New(trader internal.Trader, dto types.WalletDTO) internal.Wallet {
	return &wallet{dto: dto, trader: trader}
}

func (w *wallet) ToDTO() types.WalletDTO { return w.dto }

func (w *wallet) Currency() types.Currency {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return currency.New(w.trader, w.dto.Currency)
}

func (w *wallet) Total() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.dto.Free.Add(w.dto.Locked)
}

func (w *wallet) Free() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.dto.Free
}

func (w *wallet) ID() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.dto.ID
}
func (w *wallet) Locked() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.dto.Locked
}

func (w *wallet) Reserved() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.dto.Reserved
}

func (w *wallet) Available() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.dto.Free.Sub(w.dto.Reserved)
}

func (w *wallet) Release(amount decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.dto.Reserved.LessThan(amount) {
		return errors.New("not enough reserved funds to release")
	}

	w.dto.Reserved = w.dto.Reserved.Sub(amount)
	return nil
}
func (w *wallet) Reserve(amount decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.dto.Free.Sub(w.dto.Reserved).LessThan(amount) {
		return errors.New("not enough available funds to freeze")
	}

	w.dto.Reserved = w.dto.Reserved.Add(amount)
	return nil
}

func (w *wallet) Update(dto types.WalletDTO) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Skip reserved
	w.dto.Free = dto.Free
	w.dto.Locked = dto.Locked
}
