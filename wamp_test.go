package poloniex

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/jcelliott/turnpike.v2"
)

func newTestWebsocketServer(t *testing.T) (*turnpike.WebsocketServer, *httptest.Server) {
	wampServer := turnpike.NewBasicWebsocketServer(wampRealm)
	httpServer := httptest.NewTLSServer(wampServer)
	return wampServer, httpServer
}

func TestNewWampClient(t *testing.T) {
	Convey("Given fake WAMP server", t, func() {
		wampServer, httpServer := newTestWebsocketServer(t)
		defer func() {
			wampServer.Close()
			httpServer.Close()
		}()

		localClient, err := wampServer.GetLocalClient(wampRealm, nil)
		if err != nil {
			t.Fatalf("Error getting local client: %s", err)
		}
		defer localClient.Close()

		client, err := NewWampClient(&tls.Config{InsecureSkipVerify: true}, func(network, addr string) (net.Conn, error) {
			So(network, ShouldEqual, "tcp")
			So(addr, ShouldEqual, "api.poloniex.com:443")
			url, err := url.Parse(httpServer.URL)
			if err != nil {
				t.Fatal(err)
			}
			return net.Dial("tcp", fmt.Sprintf("localhost:%s", url.Port()))
		})
		defer client.Close()

		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}

		messageChan := make(chan interface{})
		errChan := make(chan error)

		if err := client.SubscribeToPair("BTC_ETH", messageChan, errChan); err != nil {
			t.Fatalf("Error subscribing to chat channel: %s", err)
		}

		Convey("It should recieve modify message", func() {
			modifyMarketMessage := map[string]interface{}{
				"type": messageTypeOrderBookModify,
				"data": map[string]interface{}{
					"type":   "bid",
					"rate":   "0.08529432",
					"amount": "5.00000000",
				},
			}

			args := []interface{}{interface{}(modifyMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("Error publising to pair: %s", err)
			}

			message := <-messageChan
			orderModify, ok := message.(OrderModify)

			So(ok, ShouldBeTrue)
			So(orderModify.Type, ShouldEqual, OrderUpdateTypeBid)
		})

		Convey("It should recieve remove message", func() {
			modifyMarketMessage := map[string]interface{}{
				"type": messageTypeOrderBookRemove,
				"data": map[string]interface{}{
					"type": "ask",
					"rate": "0.08529432",
				},
			}

			args := []interface{}{interface{}(modifyMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("Error publising to pair: %s", err)
			}

			message := <-messageChan
			orderRemove, ok := message.(OrderRemove)

			So(ok, ShouldBeTrue)
			So(orderRemove.Type, ShouldEqual, OrderUpdateTypeAsk)
		})

		Convey("It should recieve trade message", func() {
			newTradeMarketMessage := map[string]interface{}{
				"type": messageTypeNewTrade,
				"data": map[string]interface{}{
					"total":   "0.01190544",
					"tradeID": "30132092",
					"type":    "sell",
					"amount":  "0.14080956",
					"date":    "2017-07-13 17:33:31",
					"rate":    "0.08455001",
				},
			}

			args := []interface{}{interface{}(newTradeMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("Error publising to pair: %s", err)
			}

			message := <-messageChan
			newTrade, ok := message.(NewTrade)

			So(ok, ShouldBeTrue)
			So(newTrade.Type, ShouldEqual, TypeSell)
		})

		Convey("It should recieve errors on sequence", func() {
			args := []interface{}{}
			kwargs := map[string]interface{}{}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("Error publising to pair: %s", err)
			}

			err := <-errChan

			So(err.Error(), ShouldEqual, `key 'seq' was not found in kwargs: map[]`)

			kwargs = map[string]interface{}{"seq": ""}
			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("Error publising to pair: %s", err)
			}

			err = <-errChan

			So(err.Error(), ShouldEqual, `sequence value ("") type is string, expected float64`)
		})

		Convey("It should recieve error on structure decode", func() {
			badMarketMessage := map[string]interface{}{
				"type": float64(1),
			}

			args := []interface{}{interface{}(badMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("Error publising to pair: %s", err)
			}

			err := <-errChan

			So(err.Error(), ShouldContainSubstring, "1 error(s) decoding")
		})
	})
}
