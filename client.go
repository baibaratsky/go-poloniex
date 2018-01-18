package poloniex

import (
	"strconv"
	"time"

	"net/http"

	"golang.org/x/time/rate"
	"gopkg.in/resty.v0"
)

const (
	defaultTimeout       = 130 * time.Second
	maxRequestsPerSecond = 6
)

type Key struct {
	Key    string
	Secret string
}

type Client struct {
	keyPool keyPool
	resty   *resty.Client
	limiter *rate.Limiter

	noncePool chan uint64
}

func NewClient(keys []Key) *Client {
	client := Client{
		keyPool: keyPool{
			keys: make(chan *Key, len(keys)),
		},
		resty:     resty.DefaultClient.SetTimeout(defaultTimeout),
		limiter:   rate.NewLimiter(maxRequestsPerSecond, 1),
		noncePool: make(chan uint64),
	}

	go func() {
		client.noncePool <- uint64(time.Now().Unix())
	}()

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

func (client *Client) nonce() string {
	nonce := <-client.noncePool
	go func() {
		client.noncePool <- nonce + 1
	}()
	return strconv.FormatUint(nonce, 10)
}

type Params map[string]string
