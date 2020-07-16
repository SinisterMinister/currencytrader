package coinbase

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("coinbase.websocketURL", "wss://ws-feed.pro.coinbase.com")
	viper.SetDefault("coinbase.websocket.incomingDataBufferSize", 1024)
	viper.SetDefault("coinbase.websocket.incomingSubscriptionBufferSize", 8)

	viper.SetDefault("coinbase.websocket.tickerHandlerInputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerInputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerInputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerInputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerInputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerInputBufferSize", 4)

	viper.SetDefault("coinbase.websocket.tickerHandlerOutputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerOutputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerOutputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerOutputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerOutputBufferSize", 4)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerOutputBufferSize", 4)

	viper.SetDefault("coinbase.streams.tickerStreamBufferSize", 8)
	viper.SetDefault("coinbase.streams.orderStreamBufferSize", 8)
}
