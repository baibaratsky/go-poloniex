package poloniex

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	Convey("Given new client", t, func() {
		keys := []Key{{"key", "secret"}}

		c := NewClient(keys)

		key := c.keyPool.Get()
		So(key.Key, ShouldEqual, "key")
		So(c.limiter.Limit(), ShouldEqual, maxRequestsPerSecond)

		Convey("Should SetTimeout", func() {
			c.SetTimeout(defaultTimeout)
		})

		Convey("Should SetTransport", func() {
			c.SetTransport(&http.Transport{})
		})

		Convey("Should SetRequestRateLimit", func() {
			c.SetRequestRateLimit(888)
			So(c.limiter.Limit(), ShouldEqual, 888)
		})
	})
}
