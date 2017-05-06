package poloniex

import (
	"time"
	"gopkg.in/resty.v0"
	"sync"
	"golang.org/x/time/rate"
)

const defaultTimeout = 10 * time.Second
const maxRequestsPerSecond = 6

type Client struct {
	key        string
	secret     string
	resty      *resty.Client
	nonceMutex sync.Mutex
	limiter    *rate.Limiter
}

func NewClient(key, secret string) *Client {
	return &Client{
		key: key,
		secret: secret,
		resty: resty.DefaultClient.SetTimeout(defaultTimeout),
		limiter: rate.NewLimiter(maxRequestsPerSecond, 1),
	}
}

func (client *Client) SetTimeout(timeout time.Duration) {
	client.resty.SetTimeout(timeout)
}

type Params map[string]string
