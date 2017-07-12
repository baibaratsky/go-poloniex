package poloniex

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
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
		So(order.Total.Equal(decimal.New(10775, -8)), ShouldBeTrue)
	})
}

func TestClientOrderBook(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"asks":[[0.00007600,1164],[0.00007620,1300]], "bids":[[0.00006901,200]], "isFrozen": 0, "seq": 18849}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{})
		client.resty.SetTransport(transportForTesting(server))

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

func TestClientOrderBookAll(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC_NXT": {"asks":[[0.00007600,1164],[0.00007620,1300]], "bids":[[0.00006901,200]], "isFrozen": 0, "seq": 18849}}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{})
		client.resty.SetTransport(transportForTesting(server))

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
