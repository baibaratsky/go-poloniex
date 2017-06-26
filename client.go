package poloniex

import (
	"time"
	"gopkg.in/resty.v0"
	"golang.org/x/time/rate"
)

const defaultTimeout = 130 * time.Second
const maxRequestsPerSecond = 7

type Key struct {
	Key    string
	Secret string
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

type Params map[string]string
