package coinbase

import (
	"strings"
	"sync"
	"time"

	"github.com/go-playground/log/v7"

	"github.com/shopspring/decimal"

	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/sinisterminister/currencytrader/types"
	providerclient "github.com/sinisterminister/currencytrader/types/provider/coinbase/client"
)

type provider struct {
	streamSvc *streamSvc

	mutex         sync.Mutex
	client        *providerclient.Client
	currencies    map[string]types.CurrencyDTO
	socketStreams map[string]chan interface{}
	accounts      map[string]string
}

func New(stop <-chan bool, client *providerclient.Client) types.Provider {
	// Instantiate websocket handler
	wssvc, err := newWebsocketSvc(stop)
	if err != nil {

	}

	// Instantiate stream service
	svc := newStreamService(stop, wssvc)
	provider := &provider{
		client:     client,
		currencies: make(map[string]types.CurrencyDTO),
		streamSvc:  svc,
		accounts:   make(map[string]string),
	}

	provider.refreshCaches()

	return provider
}

func (p *provider) AttemptOrder(req types.OrderRequestDTO) (dto types.OrderDTO, err error) {
	orderRequest := coinbasepro.Order{
		Price:     req.Price.String(),
		Size:      req.Quantity.String(),
		Side:      strings.ToLower(string(req.Side)),
		ProductID: req.Market.Name,
		PostOnly:  req.ForceMaker,
	}

	placedOrder, err := p.client.CreateOrder(&orderRequest)
	if err != nil {
		return
	}

	// Convert the order to a DTO
	dto = types.OrderDTO{
		Request:      req,
		Market:       req.Market,
		CreationTime: time.Time(placedOrder.CreatedAt),
		Filled:       decimal.RequireFromString(placedOrder.FilledSize),
		ID:           placedOrder.ID,
		Status:       getStatus(placedOrder),
	}
	return
}

func (p *provider) AverageTradeVolume(mkt types.MarketDTO) (decimal.Decimal, error) {
	var trades, buffer []coinbasepro.Trade
	trades = []coinbasepro.Trade{}

	// Get the trades
	cursor := p.client.ListTrades(mkt.Name)
	for cursor.HasMore {
		if err := cursor.NextPage(&buffer); err != nil {
			for _, t := range trades {
				trades = append(trades, t)
			}
		}
	}

	// Get the average volume for the trades
	avg := decimal.Zero
	for i, t := range trades {
		avg = avg.Mul(decimal.NewFromFloat(float64(i))).Add(decimal.RequireFromString(t.Size)).Div(decimal.NewFromFloat(float64(i + 1)))
	}

	return avg, nil
}

func (p *provider) CancelOrder(ord types.OrderDTO) (err error) {
	err = p.client.CancelOrder(ord.ID)
	return
}

func (p *provider) Candles(mkt types.MarketDTO, interval types.CandleInterval, start time.Time, end time.Time) (candles []types.CandleDTO, err error) {
	// Convert the interval into a granularity
	granularity, err := time.ParseDuration(string(interval))
	if err != nil {
		return nil, err
	}

	// Create a slice for candles
	candles = []types.CandleDTO{}

	// Get the rates from the server
	rates, err := p.client.GetHistoricRates(mkt.Name, coinbasepro.GetHistoricRatesParams{
		Start:       start,
		End:         end,
		Granularity: int(granularity.Seconds()),
	})

	// Convert them into CandleDTOs
	for _, rate := range rates {
		candles = append(candles, types.CandleDTO{
			Close:     decimal.NewFromFloat(rate.Close),
			Open:      decimal.NewFromFloat(rate.Open),
			High:      decimal.NewFromFloat(rate.High),
			Low:       decimal.NewFromFloat(rate.Low),
			Volume:    decimal.NewFromFloat(rate.Volume),
			Timestamp: rate.Time,
		})
	}

	return
}

func (p *provider) Currencies() (curs []types.CurrencyDTO, err error) {
	rawCurs, err := p.client.GetCurrencies()
	if err != nil {
		return
	}

	curs = []types.CurrencyDTO{}

	for _, rc := range rawCurs {
		curs = append(curs, types.CurrencyDTO{
			Name:      rc.Name,
			Symbol:    rc.ID,
			Precision: strings.Index(rc.MinSize, "1") - 1,
			Increment: decimal.RequireFromString(rc.MinSize),
		})
	}
	return
}

func (p *provider) Fees() (fees types.FeesDTO, err error) {
	rawFees, err := p.client.GetFees()
	if err != nil {
		return
	}

	fees = types.FeesDTO{
		MakerRate: rawFees.MakerRate,
		TakerRate: rawFees.TakerRate,
		Volume:    rawFees.Volume,
	}
	return
}

func (p *provider) Markets() (mkts []types.MarketDTO, err error) {
	products, err := p.client.GetProducts()

	mkts = []types.MarketDTO{}

	for _, product := range products {

		mkts = append(mkts, types.MarketDTO{
			Name:             product.ID,
			BaseCurrency:     p.getCurrency(product.BaseCurrency),
			QuoteCurrency:    p.getCurrency(product.QuoteCurrency),
			MinPrice:         decimal.RequireFromString(product.QuoteIncrement),
			MaxPrice:         decimal.Zero,
			PriceIncrement:   decimal.RequireFromString(product.QuoteIncrement),
			MinQuantity:      decimal.RequireFromString(product.BaseMinSize),
			MaxQuantity:      decimal.RequireFromString(product.BaseMaxSize),
			QuantityStepSize: p.getCurrency(product.BaseCurrency).Increment,
		})
	}

	return
}

func (p *provider) Order(market types.MarketDTO, id string) (ord types.OrderDTO, err error) {
	raw, err := p.client.GetOrder(id)
	if err != nil {
		return
	}

	ord.CreationTime = time.Time(raw.CreatedAt)
	ord.Filled = decimal.RequireFromString(raw.FilledSize)
	ord.ID = raw.ID
	ord.Status = getStatus(raw)
	ord.Request = types.OrderRequestDTO{
		Market:   market,
		Type:     getType(raw),
		Side:     getSide(raw),
		Price:    decimal.RequireFromString(raw.Price),
		Quantity: decimal.RequireFromString(raw.Size),
	}
	ord.Market = market
	ord.Fees = decimal.RequireFromString(raw.FillFees)
	if raw.Funds != "" {
		ord.Paid = decimal.RequireFromString(raw.Funds).Add(ord.Fees)
	} else {
		ord.Paid = decimal.Zero
	}
	return
}

func (p *provider) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	return p.streamSvc.OrderStream(stop, order)
}

func (p *provider) Ticker(market types.MarketDTO) (tkr types.TickerDTO, err error) {
	raw, err := p.client.GetTicker(market.Name)
	if err != nil {
		return
	}

	tkr.Ask = decimal.RequireFromString(raw.Ask)
	tkr.Bid = decimal.RequireFromString(raw.Bid)
	tkr.Price = decimal.RequireFromString(raw.Price)
	tkr.Quantity = decimal.RequireFromString(raw.Size)
	tkr.Timestamp = time.Time(raw.Time)
	tkr.Volume = decimal.RequireFromString(string(raw.Volume))
	return
}

func (p *provider) TickerStream(stop <-chan bool, market types.MarketDTO) (stream <-chan types.TickerDTO, err error) {
	return p.streamSvc.TickerStream(stop, market)
}

func (p *provider) Wallet(currency types.CurrencyDTO) (wal types.WalletDTO, err error) {
	acct, err := p.client.GetAccount(p.accounts[currency.Symbol])
	if err != nil {
		return
	}

	wal = types.WalletDTO{
		Currency: p.getCurrency(acct.Currency),
		Free:     decimal.RequireFromString(acct.Available),
		Locked:   decimal.RequireFromString(acct.Hold),
	}
	return
}

func (p *provider) Wallets() (wals []types.WalletDTO, err error) {
	accts, err := p.client.GetAccounts()
	if err != nil {
		return
	}

	wals = []types.WalletDTO{}
	for _, acct := range accts {
		wals = append(wals, types.WalletDTO{
			Currency: p.getCurrency(acct.Currency),
			Free:     decimal.RequireFromString(acct.Available),
			Locked:   decimal.RequireFromString(acct.Hold),
			ID:       acct.ID,
		})
	}
	return
}

func (p *provider) refreshCaches() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Refresh currency cache
	curs, err := p.Currencies()
	if err != nil {
		log.WithError(err).Fatal("could not refresh currency cache from server")
	}
	for _, cur := range curs {
		p.currencies[cur.Symbol] = cur
	}

	accts, err := p.client.GetAccounts()
	if err != nil {
		log.WithError(err).Error("Failed fetching accounts")
		return
	}
	for _, acct := range accts {
		p.accounts[acct.Currency] = acct.ID
	}
}

func (p *provider) getCurrency(symbol string) (c types.CurrencyDTO) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.currencies[symbol]
}
