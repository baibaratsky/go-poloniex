package poloniex

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
	resty "gopkg.in/resty.v0"
)

type fakeHandler struct {
	HandleFunc func(w http.ResponseWriter, r *http.Request)
}

func (h *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandleFunc(w, r)
}

func createFakeServer(h *fakeHandler) *httptest.Server {
	server := httptest.NewTLSServer(h)

	return server
}

func transportForTesting(server *httptest.Server) *http.Transport {
	return &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("tcp", server.URL[strings.LastIndex(server.URL, "/")+1:])
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func TestOrder(t *testing.T) {
	Convey("UnmarshallJSON method", t, func() {
		data := []byte("[0.00300888, 0.03580906]")
		order := Order{}
		order.UnmarshalJSON(data)

		So(order.Rate.Equal(decimal.New(300888, -8)), ShouldBeTrue)
		So(order.Amount.Equal(decimal.New(3580906, -8)), ShouldBeTrue)
		So(order.Total.String(), ShouldEqual, decimal.New(1077451644528, -16).String())
	})
}

func TestClient_OrderBook(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"asks":[[0.00007600,1164],[0.00007620,1300]], "bids":[[0.00006901,200]], "isFrozen": 0, "seq": 18849}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{})
		client.SetTransport(transportForTesting(server))

		Convey("Should return OrderBook", func() {
			orderBook, err := client.OrderBook("STUBED")
			So(err, ShouldBeNil)
			So(len(orderBook.Asks), ShouldEqual, 2)
			So(orderBook.Asks[0].Amount.Equal(decimal.New(1164, 0)), ShouldBeTrue)
			So(len(orderBook.Bids), ShouldEqual, 1)
			So(orderBook.Bids[0].Amount.Equal(decimal.New(2, 2)), ShouldBeTrue)
			So(bool(orderBook.IsFrozen), ShouldBeFalse)
			So(orderBook.Sequence, ShouldEqual, 18849)
		})
	})
}

func TestClient_OrderBookAll(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC_NXT": {"asks":[[0.00007600,1164],[0.00007620,1300]], "bids":[[0.00006901,200]], "isFrozen": 0, "seq": 18849}}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{})
		client.SetTransport(transportForTesting(server))

		Convey("Should return OrderBooks", func() {
			orderBooks, err := client.OrderBookAll()
			So(err, ShouldBeNil)

			orderBook := orderBooks["BTC_NXT"]

			So(len(orderBook.Asks), ShouldEqual, 2)
			So(orderBook.Asks[0].Amount.Equal(decimal.New(1164, 0)), ShouldBeTrue)
			So(len(orderBook.Bids), ShouldEqual, 1)
			So(orderBook.Bids[0].Amount.Equal(decimal.New(2, 2)), ShouldBeTrue)
			So(bool(orderBook.IsFrozen), ShouldBeFalse)
			So(orderBook.Sequence, ShouldEqual, 18849)
		})
	})
}

func TestClient_Currencies(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"1CR":{"id":1,"name":"1CRedit","txFee":"0.01000000","minConf":3,"depositAddress":null,"disabled":0,"delisted":1,"frozen":0},"ABY":{"id":2,"name":"ArtByte","txFee":"0.01000000","minConf":8,"depositAddress":null,"disabled":0,"delisted":1,"frozen":0}}`)
			},
		}

		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{})
		client.SetTransport(transportForTesting(server))

		Convey("Should return currencies", func() {
			currencies, err := client.Currencies()
			So(err, ShouldBeNil)
			So(len(currencies), ShouldEqual, 2)
		})
	})
}

func TestClient_Ticker(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC_LTC":{"last":"0.0251","lowestAsk":"0.02589999","highestBid":"0.0251","percentChange":"0.02390438",
"baseVolume":"6.16485315","quoteVolume":"245.82513926"},"BTC_NXT":{"last":"0.00005730","lowestAsk":"0.00005710",
"highestBid":"0.00004903","percentChange":"0.16701570","baseVolume":"0.45347489","quoteVolume":"9094"}}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{})
		client.SetTransport(transportForTesting(server))

		Convey("Should return ticker", func() {
			ticker, err := client.Ticker()
			So(err, ShouldBeNil)

			So(len(ticker), ShouldEqual, 2)
			So(ticker["BTC_LTC"].Last.Equal(decimal.New(251, -4)), ShouldBeTrue)
		})
	})
}

func TestClient_publicApiRequest(t *testing.T) {
	type fields struct {
		resty      *resty.Client
		limiter    *rate.Limiter
		handleFunc http.HandlerFunc
	}
	type args struct {
		result interface{}
		method string
		params []Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "too much arguments",
			fields: fields{
				resty:      resty.DefaultClient,
				limiter:    nil,
				handleFunc: func(w http.ResponseWriter, r *http.Request) {},
			},
			args: args{
				result: make(map[string]string),
				method: "any",
				params: []Params{
					Params{"first": "first"},
					Params{"second": "second"},
				},
			},
			wantErr: true,
		},
		{
			name: "limiter error",
			fields: fields{
				resty:      resty.DefaultClient,
				limiter:    rate.NewLimiter(1, 0),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {},
			},
			args: args{
				result: make(map[string]string),
				method: "any",
				params: []Params{
					Params{"first": "first"},
				},
			},
			wantErr: true,
		},
		{
			name: "unmarshal error",
			fields: fields{
				resty:   resty.DefaultClient,
				limiter: rate.NewLimiter(1, 1),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, `///`)
				},
			},
			args: args{
				result: make(map[string]string),
				method: "ERROR",
				params: []Params{},
			},
			wantErr: true,
		},
		{
			name: "response error",
			fields: fields{
				resty:   resty.DefaultClient,
				limiter: rate.NewLimiter(1, 1),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, `{"error": "some"}`)
				},
			},
			args: args{
				result: make(map[string]string),
				method: "ERROR",
				params: []Params{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &fakeHandler{
				HandleFunc: tt.fields.handleFunc,
			}
			server := createFakeServer(handler)
			defer server.Close()
			client := &Client{
				resty:   tt.fields.resty,
				limiter: tt.fields.limiter,
			}
			transport := transportForTesting(server)
			client.SetTransport(transport)
			if err := client.publicApiRequest(tt.args.result, tt.args.method, tt.args.params...); (err != nil) != tt.wantErr {
				t.Errorf("Client.publicApiRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
