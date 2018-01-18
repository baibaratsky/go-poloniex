package poloniex

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"gopkg.in/resty.v0"
)

const (
	defaultTimeout       = 130 * time.Second
	maxRequestsPerSecond = 6
)

// Key holds data about api key.
// Use NewKey as initializer.
type Key struct {
	key    string
	secret string
	nonce  uint64
}

// NewKey returns initialized key.
func NewKey(key, secret string) Key {
	return Key{
		key:    key,
		secret: secret,
		nonce:  uint64(time.Now().Unix()),
	}
}

type Client struct {
	keyPool keyPool
	resty   *resty.Client
	limiter *rate.Limiter
}

func NewClient(keys []Key) *Client {
	client := Client{
		keyPool: keyPool{
			keys: make(chan *Key, len(keys)),
		},
		resty:   resty.DefaultClient.SetTimeout(defaultTimeout),
		limiter: rate.NewLimiter(maxRequestsPerSecond, 1),
	}

	for i := range keys {
		client.keyPool.Put(&keys[i])
	}

	return &client
}

func (client *Client) SetTimeout(timeout time.Duration) {
	client.resty.SetTimeout(timeout)
}

func (client *Client) SetTransport(transport *http.Transport) {
	client.resty.SetTransport(transport)
}

func (client *Client) SetRequestRateLimit(limit rate.Limit) {
	client.limiter.SetLimit(limit)
}

type Params map[string]string
