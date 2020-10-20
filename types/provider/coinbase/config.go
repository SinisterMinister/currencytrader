package coinbase

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("coinbase.websocketURL", "wss://ws-feed.pro.coinbase.com")
	viper.SetDefault("coinbase.websocket.incomingDataBufferSize", 1024)
	viper.SetDefault("coinbase.websocket.incomingSubscriptionBufferSize", 8)

	viper.SetDefault("coinbase.websocket.tickerHandlerInputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerInputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerInputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerInputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerInputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerInputBufferSize", 32)

	viper.SetDefault("coinbase.websocket.tickerHandlerOutputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerOutputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerOutputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerOutputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerOutputBufferSize", 32)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerOutputBufferSize", 32)

	viper.SetDefault("coinbase.streams.tickerStreamBufferSize", 64)
	viper.SetDefault("coinbase.streams.orderStreamBufferSize", 8)
}
