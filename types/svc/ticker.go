package svc

import (
	"github.com/sinisterminister/moneytrader/types"
	"github.com/sirupsen/logrus"
)

type Ticker struct {
	provider types.Provider
	streams  map[types.Market][]*streamWrapper
	sources  map[types.Market]*sourceWrapper
	running  bool
	stop     chan bool
}

type streamWrapper struct {
	market types.Market
	stream chan types.Ticker
	stop   <-chan bool
}

type sourceWrapper struct {
	market types.Market
	stream <-chan types.Ticker
	stop   chan bool
}

func NewTicker(provider types.Provider) types.TickerSvc {
	svc := &Ticker{
		provider: provider,
		streams:  make(map[types.Market][]*streamWrapper),
	}

	return svc
}

func (t *Ticker) Ticker(market types.Market) (types.Ticker, error) {
	return t.provider.GetTicker(market)
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
	return stream
}

func (t *Ticker) registerStream(wrapper *streamWrapper) {
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

func (t *Ticker) handleSource(market types.Market) *sourceWrapper {
	stop := make(chan bool)
	stream, err := t.provider.GetTickerStream(stop, market)
	wrapper := &sourceWrapper{
		stop:   stop,
		stream: stream,
		market: market,
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
			case data := <-wrapper.stream:
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
		wrapper.stream <- data
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
	t.running = true
	t.stop = make(chan bool)
	t.refreshSources()
}

func (t *Ticker) Stop() {
	close(t.stop)
	t.shutdownStreams()
	t.running = false
}
