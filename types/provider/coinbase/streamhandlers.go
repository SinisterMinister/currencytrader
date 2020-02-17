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

type orderOpenHandler struct {
	input  chan DataPackage
	output chan Open

	log log.Entry
}

func newOrderOpenHandler(stop <-chan bool) *orderOpenHandler {
	handler := &orderOpenHandler{
		input:  make(chan DataPackage),
		output: make(chan Open),
		log:    log.WithField("source", "coinbase.orderOpenHandler"),
	}

	go handler.process(stop)

	return handler
}

func (h *orderOpenHandler) Input() chan<- DataPackage {
	return h.input
}

func (h *orderOpenHandler) Output() <-chan Open {
	return h.output
}

func (h *orderOpenHandler) process(stop <-chan bool) {
	h.log.Debug("starting order open handler")
	for {
		select {
		case <-stop:
			// Time to stop
			h.log.Debug("stopping order open handler")
			return
		case pkg := <-h.input:
			// Process data
			var order Open
			h.log.Debug("parsing order open data")
			if err := json.Unmarshal(pkg.Data, &order); err != nil {
				h.log.WithError(err).Error("could not parse order open data")
			}

			h.log.Debug("sending order open data")
			h.output <- order
		}
	}
}

type orderDoneHandler struct {
	input  chan DataPackage
	output chan Done

	log log.Entry
}

func newOrderDoneHandler(stop <-chan bool) *orderDoneHandler {
	handler := &orderDoneHandler{
		input:  make(chan DataPackage),
		output: make(chan Done),
		log:    log.WithField("source", "coinbase.orderDoneHandler"),
	}

	go handler.process(stop)

	return handler
}

func (h *orderDoneHandler) Input() chan<- DataPackage {
	return h.input
}

func (h *orderDoneHandler) Output() <-chan Done {
	return h.output
}

func (h *orderDoneHandler) process(stop <-chan bool) {
	h.log.Debug("starting order done handler")
	for {
		select {
		case <-stop:
			// Time to stop
			h.log.Debug("stopping order done handler")
			return
		case pkg := <-h.input:
			// Process data
			var order Done
			h.log.Debug("parsing order done data")
			if err := json.Unmarshal(pkg.Data, &order); err != nil {
				h.log.WithError(err).Error("could not parse order done data")
			}

			h.log.Debug("sending order done data")
			h.output <- order
		}
	}
}

type orderMatchHandler struct {
	input  chan DataPackage
	output chan Match

	log log.Entry
}

func newOrderMatchHandler(stop <-chan bool) *orderMatchHandler {
	handler := &orderMatchHandler{
		input:  make(chan DataPackage),
		output: make(chan Done),
		log:    log.WithField("source", "coinbase.orderMatchHandler"),
	}

	go handler.process(stop)

	return handler
}

func (h *orderMatchHandler) Input() chan<- DataPackage {
	return h.input
}

func (h *orderMatchHandler) Output() <-chan Match {
	return h.output
}

func (h *orderMatchHandler) process(stop <-chan bool) {
	h.log.Debug("starting order match handler")
	for {
		select {
		case <-stop:
			// Time to stop
			h.log.Debug("stopping order match handler")
			return
		case pkg := <-h.input:
			// Process data
			var order Match
			h.log.Debug("parsing order match data")
			if err := json.Unmarshal(pkg.Data, &order); err != nil {
				h.log.WithError(err).Error("could not parse order match data")
			}

			h.log.Debug("sending order match data")
			h.output <- order
		}
	}
}
