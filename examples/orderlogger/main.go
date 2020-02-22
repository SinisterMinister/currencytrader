package main

import (
	"os"
	"os/signal"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types/order"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase"
	"github.com/spf13/viper"

	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
)

func main() {
	// Setup the console logger
	log.AddHandler(console.New(true), log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.NoticeLevel, log.FatalLevel, log.AlertLevel, log.PanicLevel)

	// Setup the kill switch
	killSwitch := make(chan bool)

	// Setup a coinbase client
	client := coinbasepro.NewClient()

	// Connect to sandbox
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    "https://api-public.sandbox.pro.coinbase.com",
		Key:        "db983743c2fa020a17502a111657b551",
		Passphrase: "throwback",
		Secret:     "SrHvi/n9HAcEoe/JXsaZlfok4O/hXULiK4OhoANFN5GS0odp5ciho1w1jmMXlQ40Br8G8GU6WGRPClmbQnUyEQ==",
	})

	// Setup sandbox websocket url
	viper.Set("coinbase.websocket.url", "wss://ws-feed-public.sandbox.pro.coinbase.com")

	// Start up a coinbase provider
	provider := coinbase.New(killSwitch, client)

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Get the available markets
	markets := trader.MarketSvc().Markets()

	// Stream the tickers to output log
	for _, mkt := range markets {
		if mkt.Name() == "BTC-USD" {
			go placeOrders(killSwitch, mkt)
		}
	}

	// Intercept the interrupt signal and pass it along
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Wait for the interrupt
	<-interrupt

	// Let the user know what happened
	log.Warn("Received an interrupt signal! Shutting down!")

	// Shutdown the trader
	trader.Stop()

	// Kill the provider
	close(killSwitch)
}

func placeOrders(stop <-chan bool, mkt types.Market) {
	// Get a ticker to get a price
	log.Infof("fetching ticker for %s", mkt.Name())
	ticker, err := mkt.Ticker()

	// Bail on error
	if err != nil {
		log.WithError(err).Error("could not get ticker")
		return
	}

	log.WithTrace().Infof("ticker for market %s is %s", mkt.Name(), ticker.ToDTO())

	// Place the order
	log.Infof("attempting to buy order")
	buy, err := mkt.AttemptOrder(order.Limit, order.Buy, ticker.Price(), decimal.NewFromFloat(.1))
	if err != nil {
		// Bail on error
		log.WithError(err).Error("could not place order")
		return
	}
	log.WithTrace().WithField("order", buy).Infof("order successfully placed")

	// Grab the order update stream
	log.Info("grabbing status stream for order")
	stream := buy.StatusStream(stop)

	log.Info("waiting for updates")
	// Watch the stream for the order status
	for status := range stream {
		log.Infof("order %s status %s", buy.ID(), status)
	}

	log.Infof("order %s finished with a status of %s", buy.ID(), buy.Status())
	p, err := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
}
