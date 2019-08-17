package internal

import "github.com/sinisterminister/currencytrader/types"

type MarketSvc interface {
	types.MarketSvc
}

type TickerSvc interface {
	types.TickerSvc
	types.Administerable
}

type WalletSvc interface {
	types.WalletSvc
	types.Administerable
}

type Wallet interface {
	types.Wallet
	UpdateWallet(dto types.WalletDTO)
}

type Trader interface {
	types.Administerable
	OrderSvc() types.OrderSvc
	WalletSvc() types.WalletSvc
	MarketSvc() types.MarketSvc
	TickerSvc() types.TickerSvc
	Provider() types.Provider
}
