package binance

import (
	"github.com/shopspring/decimal"

	"github.com/sinisterminister/moneytrader/pkg/wallet"

	"github.com/sinisterminister/moneytrader/pkg/currency"
	"github.com/sinisterminister/moneytrader/pkg/market"
	"github.com/sinisterminister/moneytrader/pkg/provider/binance/api"
)

type Provider struct {
	stopChan <-chan bool
}

func newProvider(stopChan <-chan bool) *Provider {
	p := &Provider{
		stopChan: stopChan,
	}

	return p
}

func (p *Provider) GetMarkets() (markets []market.Market, err error) {
	markets = []market.Market{}
	symbols := api.GetExchangeInfo().Symbols

	for _, symbol := range symbols {
		baseCur := currency.Currency{
			Name:      symbol.BaseAsset,
			Symbol:    symbol.BaseAsset,
			Precision: symbol.BasePrecision,
		}
		quoteCur := currency.Currency{
			Name:      symbol.QuoteAsset,
			Symbol:    symbol.QuoteAsset,
			Precision: symbol.QuotePrecision,
		}

		m := market.Market{
			Name:             symbol.Symbol,
			BaseCurrency:     baseCur,
			QuoteCurrency:    quoteCur,
			MinPrice:         symbol.Filters.Price.MinPrice,
			MaxPrice:         symbol.Filters.Price.MaxPrice,
			PriceIncrement:   symbol.Filters.Price.TickSize,
			MinQuantity:      symbol.Filters.LotSize.MinQuantity,
			MaxQuantity:      symbol.Filters.LotSize.MaxQuantity,
			QuantityStepSize: symbol.Filters.LotSize.StepSize,
		}

		markets = append(markets, m)
	}

	return markets, err
}

func (p *Provider) GetCurrencies() (currencies map[string]currency.Currency, err error) {
	// First, update currencies if necessary
	symbols := api.GetExchangeInfo().Symbols
	currencies = make(map[string]currency.Currency)

	for _, symbol := range symbols {
		if _, ok := currencies[symbol.BaseAsset]; !ok {
			currencies[symbol.BaseAsset] = currency.Currency{
				Symbol:    symbol.BaseAsset,
				Name:      symbol.BaseAsset,
				Precision: symbol.BasePrecision,
			}
		}

		if _, ok := currencies[symbol.QuoteAsset]; !ok {
			currencies[symbol.QuoteAsset] = currency.Currency{
				Symbol:    symbol.QuoteAsset,
				Name:      symbol.QuoteAsset,
				Precision: symbol.QuotePrecision,
			}
		}
	}

	return currencies, err
}

func (p *Provider) GetWallets() (wallets []*wallet.Wallet, err error) {
	data, err := api.GetUserData()
	if err != nil {
		return
	}

	wallets = []*wallet.Wallet{}
	currencies, _ := p.GetCurrencies()
	for _, bal := range data.Balances {
		cur, _ := currencies[bal.Asset]
		w := wallet.NewWallet(cur, bal.Free.Add(bal.Locked), bal.Free, bal.Locked, decimal.Zero)

		wallets = append(wallets, w)
	}

	return wallets, err
}
