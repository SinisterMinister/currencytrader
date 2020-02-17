package svc

import (
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	"github.com/sinisterminister/currencytrader/types/ticker"
	"github.com/spf13/viper"
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
	stop   <-chan bool

	mutex  sync.Mutex
	stream chan types.Ticker
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
	dto, err := t.trader.Provider().Ticker(m.ToDTO())
	if err != nil {
		return
	}
	tkr = ticker.New(dto)
	return
}

func (t *Ticker) TickerStream(stop <-chan bool, market types.Market) <-chan types.Ticker {
	stream := make(chan types.Ticker, viper.GetInt("currencytrader.tickersvc.streamBufferSize"))
	wrapper := &streamWrapper{
		market: market,
		stream: stream,
		stop:   stop,
	}
	t.registerStream(wrapper)

	go func() {
		<-stop
		t.deregisterStream(wrapper)
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
				c.mutex.Lock()
				close(c.stream)
				c.mutex.Unlock()
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
	stream, err := t.trader.Provider().TickerStream(stop, mkt.ToDTO())
	wrapper := &sourceWrapper{
		stop:   stop,
		stream: stream,
		market: mkt,
	}
	go func(wrapper *sourceWrapper) {
		if err != nil {
			log.WithError(err).Errorf("Could not get stream for market %s", wrapper.market.Name())
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
				data := ticker.New(payload)
				t.broadcastToStreams(wrapper.market, data)
			}
		}
	}(wrapper)

	return wrapper
}

func (t *Ticker) broadcastToStreams(market types.Market, data types.Ticker) {
	t.mutex.RLock()
	streams, ok := t.streams[market]

	if !ok {
		// No streams to broadcast to
		t.mutex.RUnlock()
		return
	}
	streams = append([]*streamWrapper{}, streams...)
	t.mutex.RUnlock()
	for _, wrapper := range streams {
		select {
		case wrapper.stream <- data:
		default:
			log.Warn("Skipping blocked ticker channel")
		}
	}
}

func (t *Ticker) shutdownStreams() {
	t.mutex.Lock()
	for market, wrapper := range t.sources {
		// Close the stream
		close(wrapper.stop)
		delete(t.sources, market)
	}
	t.mutex.Unlock()
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
