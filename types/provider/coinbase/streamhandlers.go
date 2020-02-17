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

type orderReceivedHandler struct {
	input  chan DataPackage
	output chan Received

	log log.Entry
}

func newOrderReceivedHandler(stop <-chan bool) *orderReceivedHandler {
	handler := &orderReceivedHandler{
		input:  make(chan DataPackage),
		output: make(chan Received),
		log:    log.WithField("source", "coinbase.orderReceivedHandler"),
	}

	go handler.process(stop)

	return handler
}

func (h *orderReceivedHandler) Input() chan<- DataPackage {
	return h.input
}

func (h *orderReceivedHandler) Output() <-chan Received {
	return h.output
}

func (h *orderReceivedHandler) process(stop <-chan bool) {
	h.log.Debug("starting order receieved handler")
	for {
		select {
		case <-stop:
			// Time to stop
			h.log.Debug("stopping order receieved handler")
			return
		case pkg := <-h.input:
			// Process data
			var order Received
			h.log.Debug("parsing order receieved data")
			if err := json.Unmarshal(pkg.Data, &order); err != nil {
				h.log.WithError(err).Error("could not parse order receieved data")
			}

			h.log.Debug("sending order receieved data")
			h.output <- order
		}
	}
}
