package svc

import (
	"github.com/sinisterminister/currencytrader/types/currency"
	"github.com/sinisterminister/currencytrader/types/wallet"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

type walletSvc struct {
	trader internal.Trader
}

func NewWallet(trader internal.Trader) internal.WalletSvc {
	return &walletSvc{
		trader: trader,
	}
}

func (w *walletSvc) Currency(name string) (currency types.Currency, err error) {
	currencies, err := w.Currencies()
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

func (w *walletSvc) Currencies() (currencies []types.Currency, err error) {
	dtos, err := w.trader.Provider().Currencies()
	if err != nil {
		return
	}

	// Convert the currencies
	currencies = []types.Currency{}
	for _, dto := range dtos {
		currencies = append(currencies, currency.New(w.trader, dto))
	}

	return
}

func (w *walletSvc) Wallet(currency types.Currency) (wal types.Wallet, err error) {
	dto, err := w.trader.Provider().Wallet(currency.ToDTO())
	if err != nil {
		return
	}

	return wallet.New(w.trader, dto), err
}

func (w *walletSvc) Wallets() (wallets []types.Wallet, err error) {
	dtos, err := w.trader.Provider().Wallets()
	if err != nil {
		return
	}

	wallets = []types.Wallet{}
	for _, dto := range dtos {
		wallets = append(wallets, wallet.New(w.trader, dto))
	}

	return
}
