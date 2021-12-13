package types

type Subscribe struct {
	Message
	AuthenticatedSubscribe
	Channels []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	} `json:"channels"`
}
