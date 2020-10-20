package coinbase

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("coinbase.websocketURL", "wss://ws-feed.pro.coinbase.com")
	viper.SetDefault("coinbase.websocket.incomingDataBufferSize", 1024)
	viper.SetDefault("coinbase.websocket.incomingSubscriptionBufferSize", 8)

	viper.SetDefault("coinbase.websocket.tickerHandlerInputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerInputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerInputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerInputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerInputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerInputBufferSize", 16)

	viper.SetDefault("coinbase.websocket.tickerHandlerOutputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerOutputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerOutputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerOutputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerOutputBufferSize", 16)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerOutputBufferSize", 16)

	viper.SetDefault("coinbase.streams.tickerStreamBufferSize", 16)
	viper.SetDefault("coinbase.streams.orderStreamBufferSize", 8)
}
