package trader

import (
	"github.com/sinisterminister/moneytrader/pkg/currency"
	"github.com/sinisterminister/moneytrader/pkg/market"
	"github.com/sinisterminister/moneytrader/pkg/provider"
	"github.com/sinisterminister/moneytrader/pkg/wallet"
)

type Trader struct {
	provider provider.Provider
}

func New(provider provider.Provider) (t *Trader) {
	t = &Trader{
		provider: provider,
	}
	return t
}

func (t *Trader) GetCurrency(symbol string) (currency.Currency, error) {
	return t.provider.GetCurrency(symbol)
}

func (t *Trader) GetCurrencies() ([]currency.Currency, error) {
	return t.provider.GetCurrencies()
}

func (t *Trader) GetMarket(cur0 currency.Currency, cur1 currency.Currency) (market market.Market, err error) {
	return t.provider.GetMarket(cur0, cur1)
}

func (t *Trader) GetAllMarkets() (markets []market.Market, err error) {
	return t.provider.GetAllMarkets()
}

func (t *Trader) GetWallet(cur currency.Currency) (wallet wallet.Wallet, err error) {
	return wallet, err
}

func (t *Trader) GetAllWallets() (wallets []wallet.Wallet, err error) {
	return wallets, err
}
