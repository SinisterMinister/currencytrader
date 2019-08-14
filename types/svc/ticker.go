package svc

import (
	"sync"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	"github.com/sinisterminister/currencytrader/types/market"
	"github.com/sinisterminister/currencytrader/types/ticker"
	"github.com/sirupsen/logrus"
)

type Ticker struct {
	trader internal.Trader

	mutex   sync.RWMutex
	streams map[types.Market][]*streamWrapper
	sources map[types.Market]*sourceWrapper
	running bool
}

type streamWrapper struct {
	market types.Market
	stream chan types.Ticker
	stop   <-chan bool
}

type sourceWrapper struct {
	market types.Market
	stream <-chan types.TickerDTO
	stop   chan bool
}

func NewTicker(trader internal.Trader) internal.TickerSvc {
	svc := &Ticker{
		trader:  trader,
		streams: make(map[types.Market][]*streamWrapper),
		sources: make(map[types.Market]*sourceWrapper),
	}

	return svc
}

func (t *Ticker) Ticker(m types.Market) (tkr types.Ticker, err error) {
	dto, err := t.trader.Provider().GetTicker(market.ToDTO(m))
	if err != nil {
		return
	}
	tkr = ticker.New(ticker.TickerConfig{dto})
	return
}

func (t *Ticker) TickerStream(stop <-chan bool, market types.Market) <-chan types.Ticker {
	stream := make(chan types.Ticker)
	wrapper := &streamWrapper{
		market: market,
		stream: stream,
		stop:   stop,
	}
	t.registerStream(wrapper)

	go func() {
		select {
		case <-stop:
			t.deregisterStream(wrapper)
		}
	}()

	t.mutex.RLock()
	r := t.running
	t.mutex.RUnlock()
	if r {
		go t.refreshSources()
	}

	return stream
}

func (t *Ticker) registerStream(wrapper *streamWrapper) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// Create slice if not already exist
	streams, ok := t.streams[wrapper.market]
	if !ok {
		streams = []*streamWrapper{}
	}

	// Add stream to registry
	streams = append(streams, wrapper)
	t.streams[wrapper.market] = streams
}

func (t *Ticker) deregisterStream(wrapper *streamWrapper) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// Bail out if there's no channels there already
	if _, ok := t.streams[wrapper.market]; !ok {
		return
	}

	filtered := t.streams[wrapper.market][:0]
	for _, c := range t.streams[wrapper.market] {
		if c != wrapper {
			filtered = append(filtered, c)
		} else {
			// Close the channel gracefully
			select {
			case <-c.stream:
			default:
				close(c.stream)
			}
		}
	}

	// Clean up references
	for i := len(filtered); i < len(t.streams[wrapper.market]); i++ {
		t.streams[wrapper.market][i] = nil
	}

	t.streams[wrapper.market] = filtered
}

func (t *Ticker) refreshSources() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// Create any missing streams
	for market := range t.streams {
		if _, ok := t.sources[market]; !ok {
			t.sources[market] = t.handleSource(market)
		}
	}

	// Kill any unneeded streams
	for market, source := range t.sources {
		if _, ok := t.streams[market]; !ok {
			// Close the stop chan
			close(source.stop)
			delete(t.sources, market)
		}
	}
}

func (t *Ticker) handleSource(mkt types.Market) *sourceWrapper {
	stop := make(chan bool)
	stream, err := t.trader.Provider().GetTickerStream(stop, market.ToDTO(mkt))
	wrapper := &sourceWrapper{
		stop:   stop,
		stream: stream,
		market: mkt,
	}
	go func(wrapper *sourceWrapper) {
		if err != nil {
			logrus.WithError(err).Errorf("Could not get stream for market %s", wrapper.market.Name())
			return
		}
		for {
			// Bail out on stop
			select {
			case <-wrapper.stop:
				return
			default:
			}

			select {
			case <-wrapper.stop:
				// Backup bailout
				return
			case payload := <-wrapper.stream:
				data := ticker.New(ticker.TickerConfig{TickerDTO: payload})
				t.broadcastToStreams(wrapper.market, data)
			}
		}
	}(wrapper)

	return wrapper
}

func (t *Ticker) broadcastToStreams(market types.Market, data types.Ticker) {
	streams, ok := t.streams[market]
	if !ok {
		// No streams to broadcast to
		return
	}
	for _, wrapper := range streams {
		select {
		case wrapper.stream <- data:
		default:
			logrus.Warn("Skipping blocked ticker channel")
		}
	}
}

func (t *Ticker) shutdownStreams() {
	for market, wrapper := range t.sources {
		// Close the stream
		close(wrapper.stop)
		delete(t.sources, market)
	}
}

func (t *Ticker) Start() {
	t.mutex.Lock()
	t.running = true
	t.mutex.Unlock()
	t.refreshSources()
}

func (t *Ticker) Stop() {
	t.mutex.Lock()
	t.running = false
	t.mutex.Unlock()
	t.shutdownStreams()
}
