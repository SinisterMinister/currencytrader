package coinbase

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/spf13/viper"
)

type tickerHandler struct {
	input  chan DataPackage
	output chan Ticker

	log log.Entry
}

func newTickerHandler(stop <-chan bool) *tickerHandler {
	handler := &tickerHandler{
		input:  make(chan DataPackage, viper.GetInt("coinbase.websocket.tickerHandlerInputBufferSize")),
		output: make(chan Ticker, viper.GetInt("coinbase.websocket.tickerHandlerOutputBufferSize")),
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

func (h *tickerHandler) Name() string {
	return "ticker"
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
			select {
			case h.output <- ticker:
			default:
				log.Warn("ticker handler output channel blocked")
			}
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
		input:  make(chan DataPackage, viper.GetInt("coinbase.websocket.orderReceivedHandlerInputBufferSize")),
		output: make(chan Received, viper.GetInt("coinbase.websocket.orderReceivedHandlerOutputBufferSize")),
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

func (h *orderReceivedHandler) Name() string {
	return "received"
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
			select {
			case h.output <- order:
			default:
				log.Warn("order received handler output channel blocked")
			}
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
		input:  make(chan DataPackage, viper.GetInt("coinbase.websocket.orderOpenHandlerInputBufferSize")),
		output: make(chan Open, viper.GetInt("coinbase.websocket.orderDoneHandlerOutputBufferSize")),
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

func (h *orderOpenHandler) Name() string {
	return "open"
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
			select {
			case h.output <- order:
			default:
				log.Warn("order open handler output channel blocked")
			}
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
		input:  make(chan DataPackage, viper.GetInt("coinbase.websocket.orderDoneHandlerInputBufferSize")),
		output: make(chan Done, viper.GetInt("coinbase.websocket.orderDoneHandlerOutputBufferSize")),
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

func (h *orderDoneHandler) Name() string {
	return "done"
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
			select {
			case h.output <- order:
			default:
				log.Warn("order done handler output channel blocked")
			}
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
		input:  make(chan DataPackage, viper.GetInt("coinbase.websocket.orderMatchHandlerInputBufferSize")),
		output: make(chan Match, viper.GetInt("coinbase.websocket.orderMatchHandlerOutputBufferSize")),
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

func (h *orderMatchHandler) Name() string {
	return "match"
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
			select {
			case h.output <- order:
			default:
				log.Warn("order match handler output channel blocked")
			}
		}
	}
}

type orderChangeHandler struct {
	input  chan DataPackage
	output chan Change

	log log.Entry
}

func newOrderChangeHandler(stop <-chan bool) *orderChangeHandler {
	handler := &orderChangeHandler{
		input:  make(chan DataPackage, viper.GetInt("coinbase.websocket.orderChangeHandlerInputBufferSize")),
		output: make(chan Change, viper.GetInt("coinbase.websocket.orderChangeHandlerOutputBufferSize")),
		log:    log.WithField("source", "coinbase.orderChangeHandler"),
	}

	go handler.process(stop)

	return handler
}

func (h *orderChangeHandler) Input() chan<- DataPackage {
	return h.input
}

func (h *orderChangeHandler) Output() <-chan Change {
	return h.output
}

func (h *orderChangeHandler) Name() string {
	return "change"
}

func (h *orderChangeHandler) process(stop <-chan bool) {
	h.log.Debug("starting order change handler")
	for {
		select {
		case <-stop:
			// Time to stop
			h.log.Debug("stopping order change handler")
			return
		case pkg := <-h.input:
			// Process data
			var order Change
			h.log.Debug("parsing order change data")
			if err := json.Unmarshal(pkg.Data, &order); err != nil {
				h.log.WithError(err).Error("could not parse order change data")
			}

			h.log.Debug("sending order change data")
			select {
			case h.output <- order:
			default:
				log.Warn("order change handler output channel blocked")
			}
		}
	}
}
