package types

type GlobalSubscribe struct {
	Message
	AuthenticatedSubscribe
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}
