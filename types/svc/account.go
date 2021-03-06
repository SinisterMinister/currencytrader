package svc

import (
	"sync"
	"time"

	"github.com/sinisterminister/currencytrader/types/currency"
	"github.com/sinisterminister/currencytrader/types/fees"
	"github.com/sinisterminister/currencytrader/types/wallet"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

type accountSvc struct {
	trader internal.Trader

	mutex    sync.Mutex
	feeCache types.Fees
	feeValid time.Time
}

func NewAccount(trader internal.Trader) internal.AccountSvc {
	return &accountSvc{
		trader: trader,
	}
}

func (svc *accountSvc) Currency(name string) (currency types.Currency, err error) {
	currencies, err := svc.Currencies()
	if err != nil {
		return
	}

	for _, cur := range currencies {
		if cur.Symbol() == name {
			return cur, nil
		}
	}
	return
}

func (svc *accountSvc) Currencies() (currencies []types.Currency, err error) {
	dtos, err := svc.trader.Provider().Currencies()
	if err != nil {
		return
	}

	// Convert the currencies
	currencies = []types.Currency{}
	for _, dto := range dtos {
		currencies = append(currencies, currency.New(svc.trader, dto))
	}

	return
}

func (svc *accountSvc) Fees() (types.Fees, error) {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()
	if svc.feeCache == nil || time.Now().After(svc.feeValid) {
		dto, err := svc.trader.Provider().Fees()
		if err != nil {
			return nil, err
		}
		svc.feeCache = fees.New(svc.trader, dto)

		// Cache the fees for 5 minutes
		svc.feeValid.Add(5 * time.Minute)
	}

	return svc.feeCache, nil
}

func (svc *accountSvc) Wallet(currency types.Currency) (wal types.Wallet, err error) {
	dto, err := svc.trader.Provider().Wallet(currency.ToDTO())
	if err != nil {
		return
	}

	return wallet.New(svc.trader, dto), err
}

func (svc *accountSvc) Wallets() (wallets []types.Wallet, err error) {
	dtos, err := svc.trader.Provider().Wallets()
	if err != nil {
		return
	}

	wallets = []types.Wallet{}
	for _, dto := range dtos {
		wallets = append(wallets, wallet.New(svc.trader, dto))
	}

	return
}
