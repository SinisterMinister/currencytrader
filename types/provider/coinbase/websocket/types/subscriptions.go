package types

type Subscriptions struct {
	Message
	Channels []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	} `json:"channels"`
}
