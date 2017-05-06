package poloniex

import (
	"time"
	"gopkg.in/resty.v0"
	"sync"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	key        string
	secret     string
	resty      *resty.Client
	nonceMutex sync.Mutex
}

func NewClient(key, secret string) *Client {
	return &Client{
		key: key,
		secret: secret,
		resty: resty.DefaultClient.SetTimeout(defaultTimeout),
	}
}

func (client *Client) SetTimeout(timeout time.Duration) {
	client.resty.SetTimeout(timeout)
}

type Params map[string]string
