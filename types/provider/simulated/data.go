package simulated

import (
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

func getCurrencies() []types.CurrencyDTO {
	return append([]types.CurrencyDTO{},
		types.CurrencyDTO{
			Name:      "US Dollar",
			Symbol:    "USD",
			Precision: 2,
		},
		types.CurrencyDTO{
			Name:      "Bitcoin",
			Symbol:    "BTC",
			Precision: 8,
		},
		types.CurrencyDTO{
			Name:      "Etherium",
			Symbol:    "ETH",
			Precision: 8,
		},
		types.CurrencyDTO{
			Name:      "Ripple",
			Symbol:    "XRP",
			Precision: 2,
		},
	)
}

func getMarkets() []types.MarketDTO {
	currencies := getCurrencies()
	markets := []types.MarketDTO{}

	contains := func(markets []types.MarketDTO, symbol string) bool {
		for _, m := range markets {
			if m.Name == symbol {
				return true
			}
		}
		return false
	}

	for _, base := range currencies {
		for _, quote := range currencies {
			if !contains(markets, base.Symbol+quote.Symbol) && !contains(markets, quote.Symbol+base.Symbol) && base.Symbol != quote.Symbol {
				markets = append(markets, types.MarketDTO{
					Name:          base.Symbol + quote.Symbol,
					BaseCurrency:  base,
					QuoteCurrency: quote,
				})
			}
		}
	}
	return markets
}

func getTicker(mkt types.MarketDTO) types.TickerDTO {
	return types.TickerDTO{
		Ask:       decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Bid:       decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Price:     decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Quantity:  decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Timestamp: time.Now(),
		Volume:    decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(10000))).Round(int32(mkt.QuoteCurrency.Precision)),
	}
}

func getTickerStream(stop <-chan bool, mkt types.MarketDTO) <-chan types.TickerDTO {
	ch := make(chan types.TickerDTO)

	go func(ch chan types.TickerDTO) {
		ticker := time.NewTicker(1 * time.Second)

		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			default:
			}

			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				ch <- getTicker(mkt)
			}
		}

	}(ch)

	return ch
}

func getWallets() []types.WalletDTO {
	currencies := getCurrencies()
	wallets := []types.WalletDTO{}
	for _, cur := range currencies {
		wallets = append(wallets,
			types.WalletDTO{
				Currency: cur,
				Free:     decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(cur.Precision)),
				Locked:   decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(cur.Precision)),
			},
		)
	}
	return wallets
}

func getWallet(cur types.CurrencyDTO) types.WalletDTO {
	return types.WalletDTO{
		Currency: cur,
		Free:     decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(cur.Precision)),
		Locked:   decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(cur.Precision)),
	}
}

func getWalletStream(stop <-chan bool, cur types.CurrencyDTO) <-chan types.WalletDTO {
	ch := make(chan types.WalletDTO)
	go func(stop <-chan bool, cur types.CurrencyDTO, ch chan types.WalletDTO) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-stop:
				return
			default:
			}

			select {
			case <-stop:
				return
			case <-ticker.C:
				ch <- getWallet(cur)
			}
		}
	}(stop, cur, ch)
	return ch
}
