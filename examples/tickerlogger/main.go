package main

import (
	"os"
	"os/signal"

	"github.com/go-playground/log/v7"
	"github.com/go-playground/log/v7/handlers/console"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase"
	"github.com/spf13/viper"
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
		go streamTicker(killSwitch, mkt)
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

func streamTicker(stop <-chan bool, market types.Market) {
	// Get the ticker stream from the market
	stream := market.TickerStream(stop)

	// Watch the stream and log any data sent over it
	for {
		// Bail out on stop
		select {
		case <-stop:
			return
		default:
		}

		select {
		//Backup bailout
		case <-stop:
			return

		// Data received
		case data := <-stream:
			if data != nil {
				log.WithField("data", data.ToDTO()).Infof("stream data received for %s market", market.Name())
			} else {
				log.Info("empty stream data received")
			}
		}
	}
}
