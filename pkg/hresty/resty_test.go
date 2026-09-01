package hresty

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRestyClient(t *testing.T) {
	Convey("NewRestyClient", t, func() {
		client := NewRestyClient()
		So(client, ShouldNotBeNil)
		So(client.GetClient().Transport, ShouldNotBeNil)
	})
}

func TestGetTraceInfo(t *testing.T) {
	Convey("GetTraceInfo", t, func() {
		Convey("nil 响应返回 error", func() {
			ti, err := GetTraceInfo(nil)
			So(err, ShouldNotBeNil)
			So(ti, ShouldBeNil)
		})

		Convey("正常请求返回 trace", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("hello"))
			}))
			defer srv.Close()

			client := NewRestyClient()
			resp, err := client.R().EnableTrace().Get(srv.URL)
			So(err, ShouldBeNil)
			So(resp.StatusCode(), ShouldEqual, http.StatusOK)

			trace, err := GetTraceInfo(resp)
			So(err, ShouldBeNil)
			So(trace, ShouldNotBeNil)
			So(trace.ResponseSize, ShouldEqual, len("hello"))
			So(trace.RemoteAddr, ShouldNotBeEmpty)
		})
	})
}
