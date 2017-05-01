package poloniex

type Client struct {
	key string
	secret string
}

func NewClient(key, secret string) *Client {
	return &Client{
		key: key,
		secret: secret,
	}
}

type Params map[string]string
