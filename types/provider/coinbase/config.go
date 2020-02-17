package coinbase

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("coinbase.websocket.url", "wss://ws-feed.pro.coinbase.com")
	viper.SetDefault("coinbase.websocket.incomingDataBufferSize", 1024)
}
