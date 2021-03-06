package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/go-playground/log/v7"

	"github.com/sinisterminister/currencytrader/types/candle"

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

	// Start a channel to stream the candles to for logging
	stream := make(chan types.CandleDTO)

	// Launch the logger
	go candleLogger(killSwitch, stream)

	// Stream the tickers to output log
	for _, mkt := range markets {
		go logCandles(killSwitch, mkt, stream)
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

func candleLogger(stop <-chan bool, stream chan types.CandleDTO) {
	for {
		select {
		case <-stop:
			return
		case candle := <-stream:
			log.Infof("candle data is %s", candle)
		}
	}
}

func logCandles(stop <-chan bool, mkt types.Market, stream chan<- types.CandleDTO) {
	// Get the candles for the market
	candles, err := mkt.Candles(candle.FiveMinutes, time.Now().Add(-5*time.Minute), time.Now())

	// Bail on error
	if err != nil {
		log.WithError(err).Error("could not get candles")
		return
	}

	for _, candle := range candles {
		select {
		case <-stop:
			// Bail out
			return
		case stream <- candle.ToDTO():
		}
	}
}
