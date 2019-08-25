package binance

type Config struct {
	WebSocketHost string `env:"BINANCE_WEBSOCKET_HOST,default=stream.binance.com"`
	WebSocketPort string `env:"BINANCE_WEBSOCKET_PORT,default=9443"`
	RestURL       string `env:"BINANCE_URL,default=https://api.binance.com"`
	SecretKey     string `env:"BINANCE_SECRET",validate:"required"`
	APIKey        string `env:"BINANCE_KEY",validate:"required"`
}
