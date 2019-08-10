package provider

import (
	"github.com/sinisterminister/moneytrader/pkg/currency"
	"github.com/sinisterminister/moneytrader/pkg/market"
	"github.com/sinisterminister/moneytrader/pkg/order"
	"github.com/sinisterminister/moneytrader/pkg/wallet"
)

type Provider interface {
	GetCurrency(symbol string) (currency.Currency, error)
	GetCurrencies() ([]currency.Currency, error)
	
	GetWallet(cur currency.Currency) (wallet.Wallet, error)
	GetWallets(cur currency.Currency) ([]wallet.Wallet, error)

	GetMarket(currency0 currency.Currency, currency1 currency.Currency) (market.Market, error)
	GetAllMarkets() ([]market.Market, error)


	// GetOrder(id string) (order.Order, error)
	// GetOrders(id ...string) ([]order.Order, error)
	// GetOpenOrders() ([]order.Order, error)
}
