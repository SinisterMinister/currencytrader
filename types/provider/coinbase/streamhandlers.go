package coinbase

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
)

type tickerHandler struct {
	input  chan DataPackage
	output chan Ticker

	log log.Entry
}

func newTickerHandler(stop <-chan bool) *tickerHandler {
	handler := &tickerHandler{
		input:  make(chan DataPackage),
		output: make(chan Ticker),
		log:    log.WithField("source", "coinbase.tickerHandler"),
	}

	go handler.process(stop)

	return handler
}

func (h *tickerHandler) Input() chan<- DataPackage {
	return h.input
}

func (h *tickerHandler) Output() <-chan Ticker {
	return h.output
}

func (h *tickerHandler) process(stop <-chan bool) {
	h.log.Debug("starting ticker handler")
	for {
		select {
		case <-stop:
			// Time to stop
			h.log.Debug("stopping ticker handler")
			return
		case pkg := <-h.input:
			// Process Ticker
			var ticker Ticker
			h.log.Debug("parsing ticker data")
			if err := json.Unmarshal(pkg.Data, &ticker); err != nil {
				h.log.WithError(err).Error("could not parse ticker data")
			}

			h.log.Debug("sending ticker data")
			h.output <- ticker
		}
	}
}
