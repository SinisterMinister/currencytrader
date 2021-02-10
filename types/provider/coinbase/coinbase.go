package coinbase

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/google/uuid"

	"github.com/shopspring/decimal"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
	providerclient "github.com/sinisterminister/currencytrader/types/provider/coinbase/client"
	"github.com/sinisterminister/go-coinbasepro/v2"
)

type provider struct {
	streamSvc *streamSvc

	mutex         sync.Mutex
	client        *providerclient.Client
	currencies    map[string]types.CurrencyDTO
	socketStreams map[string]chan interface{}
	accounts      map[string]string
	rateLimiter   chan interface{}
	throttleOnce  sync.Once
}

func New(stop <-chan bool, client *providerclient.Client, rateLimit int, burstLimit int) types.Provider {
	// Instantiate websocket handler
	wssvc, err := newWebsocketSvc(stop)
	if err != nil {

	}

	// Instantiate stream service
	svc := newStreamService(stop, wssvc)
	provider := &provider{
		client:      client,
		currencies:  make(map[string]types.CurrencyDTO),
		streamSvc:   svc,
		accounts:    make(map[string]string),
		rateLimiter: make(chan interface{}, burstLimit-rateLimit),
	}
	provider.startThrottler(rateLimit, burstLimit)
	provider.refreshCaches()

	return provider
}

func (p *provider) AttemptOrder(req types.OrderRequestDTO) (dto types.OrderDTO, err error) {
	// Create a client order id
	cid, err := uuid.NewRandom()
	if err != nil {
		return
	}

	var orderRequest coinbasepro.Order
	switch req.Type {
	case order.Limit:
		// Create the limit order from the request
		orderRequest = coinbasepro.Order{
			Price:     req.Price.String(),
			Size:      req.Quantity.String(),
			Side:      strings.ToLower(string(req.Side)),
			ProductID: req.Market.Name,
			PostOnly:  req.ForceMaker,
			ClientOID: cid.String(),
		}
	case order.Market:
		// Create the market order from the request
		var funds, size string
		if req.Funds.Equal(decimal.Zero) {
			funds = ""
		} else {
			funds = req.Funds.String()
		}
		if req.Quantity.Equal(decimal.Zero) {
			size = ""
		} else {
			size = req.Quantity.String()
		}

		orderRequest = coinbasepro.Order{
			Type:      "market",
			Funds:     funds,
			Size:      size,
			Side:      strings.ToLower(string(req.Side)),
			ProductID: req.Market.Name,
			ClientOID: cid.String(),
		}
	default:
		return types.OrderDTO{}, fmt.Errorf("order type %s not implemented", req.Type)
	}

	// Mind the rate limit
	<-p.rateLimiter

	// Place the order
	placedOrder, err := p.client.CreateOrder(&orderRequest)
	if err != nil {
		// Make sure the order didn't manage to make it there somehow
		log.WithError(err).Debugf("error creating order %s checking if it posted", cid.String())
		var err2 error
		placedOrder, err2 = p.client.GetOrder("client:" + cid.String())
		if err2 != nil {
			return
		}
	}

	if req.Type == order.Market {
		req.Price, _ = decimal.NewFromString(placedOrder.Price)
	}

	// Convert the order to a DTO
	dto = types.OrderDTO{
		Request:      req,
		Market:       req.Market,
		CreationTime: time.Time(placedOrder.CreatedAt),
		Filled:       decimal.RequireFromString(placedOrder.FilledSize),
		ID:           cid.String(),
		Status:       getStatus(placedOrder),
	}

	// Register client ID
	p.streamSvc.registerClientId(placedOrder.ID, cid.String())
	return
}

func (p *provider) AverageTradeVolume(mkt types.MarketDTO) (decimal.Decimal, error) {
	var trades, buffer []coinbasepro.Trade
	trades = []coinbasepro.Trade{}

	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter
	err = p.client.CancelOrder(fmt.Sprintf("client:%s", ord.ID))
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

	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter
	log.Debugf("getting order %s", fmt.Sprintf("client:%s", id))

	raw, err := p.client.GetOrder(fmt.Sprintf("client:%s", id))
	if err != nil {
		return
	}

	// Normalize the price, size, and funds
	price, _ := decimal.NewFromString(raw.Price)
	execVal, _ := decimal.NewFromString(raw.ExecutedValue)
	size, _ := decimal.NewFromString(raw.Size)
	funds, _ := decimal.NewFromString(raw.Funds)
	filled, _ := decimal.NewFromString(raw.FilledSize)

	// Set the price for market orders
	if price.Equal(decimal.Zero) && !execVal.Equal(decimal.Zero) && !filled.Equal(decimal.Zero) {
		price = execVal.Div(filled)
	}

	ord.CreationTime = time.Time(raw.CreatedAt)
	ord.Filled = filled
	ord.ID = id
	ord.Status = getStatus(raw)
	ord.Request = types.OrderRequestDTO{
		Market:   market,
		Type:     getType(raw),
		Side:     getSide(raw),
		Price:    price,
		Quantity: size,
		Funds:    funds,
	}
	ord.Market = market
	ord.Fees = decimal.RequireFromString(raw.FillFees)
	if raw.Funds != "" {
		ord.Paid = decimal.RequireFromString(raw.Funds).Add(ord.Fees)
	} else {
		ord.Paid = decimal.Zero
	}
	p.streamSvc.registerClientId(raw.ID, id)
	return
}

func (p *provider) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	return p.streamSvc.OrderStream(stop, order)
}

func (p *provider) RefreshOrder(in types.OrderDTO) (out types.OrderDTO, err error) {
	out, err = p.Order(in.Market, in.ID)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			log.Debugf("could not find order %s in API; assuming it was cancelled", in.ID)
			out = in
			out.Status = order.Canceled
			err = nil
		}
	}
	return
}

func (p *provider) Ticker(market types.MarketDTO) (tkr types.TickerDTO, err error) {
	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter

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
	// Mind the rate limit
	<-p.rateLimiter

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

func (p *provider) startThrottler(rateLimit int, burstLimit int) {
	p.mutex.Lock()
	limiter := p.rateLimiter
	p.mutex.Unlock()
	go func(limiter chan interface{}) {
		ticker := time.Tick(time.Second / time.Duration(rateLimit))
		for {
			<-ticker
			limiter <- struct{}{}
		}
	}(limiter)
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

	// Mind the rate limit
	<-p.rateLimiter
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
