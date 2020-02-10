package main

import (
	"os"
	"os/signal"

	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types/order"

	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/simulated"
)

func main() {
	// Start up the simulated provider
	provider := simulated.New(simulated.ProviderConfig{})

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Get the available markets
	markets := trader.MarketSvc().Markets()

	// Setup a close channel
	killSwitch := make(chan bool)

	// Stream the tickers to output log
	for _, mkt := range markets {
		go placeOrders(killSwitch, mkt)
	}

	// Intercept the interrupt signal and pass it along
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Wait for the interrupt
	<-interrupt

	// Let the user know what happened
	log.Warn("Received an interrupt signal! Shutting down!")

	// Kill the streams
	close(killSwitch)

	// Shutdown the trader
	trader.Stop()
}

func placeOrders(stop <-chan bool, mkt types.Market) {
	// Get a ticker to get a price
	ticker, err := mkt.Ticker()

	// Bail on error
	if err != nil {
		log.WithError(err).Error("could not get ticker")
		return
	}

	log.Infof("ticker for market %s is %s", mkt.Name(), ticker.ToDTO())

	// Place the order
	buy, err := mkt.AttemptOrder(order.Limit, order.Buy, ticker.Price(), decimal.NewFromFloat(10))
	if err != nil {
		// Bail on error
		log.WithError(err).Error("could not place order")
		return
	}

	// Grab the order update stream
	stream := buy.StatusStream(stop)

	// Watch the stream for the order status
	for status := range stream {
		log.Infof("order %s status %s", buy.ID(), status)
	}

	log.Infof("order %s finished with a status of %s", buy.ID(), buy.Status())
}
