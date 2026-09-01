package hlark

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenSign(t *testing.T) {
	Convey("GenSign", t, func() {
		Convey("secret 为空返回空串", func() {
			s, err := GenSign(1234567890, "")
			So(err, ShouldBeNil)
			So(s, ShouldBeEmpty)
		})

		Convey("固定 timestamp+secret 输出稳定", func() {
			s1, err := GenSign(1700000000, "my-secret")
			So(err, ShouldBeNil)
			s2, _ := GenSign(1700000000, "my-secret")
			So(s1, ShouldEqual, s2)
			So(s1, ShouldNotBeEmpty)
		})

		Convey("不同 secret 输出不同", func() {
			a, _ := GenSign(1700000000, "secret-a")
			b, _ := GenSign(1700000000, "secret-b")
			So(a, ShouldNotEqual, b)
		})
	})
}

func TestSendTextToFeiShu(t *testing.T) {
	Convey("SendTextToFeiShu", t, func() {
		Convey("响应 code=0 视为成功", func() {
			// 在 handler 协程中只做捕获,断言放回主协程,避免 goconvey 上下文丢失
			var gotReq feiShuRequest
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &gotReq)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
			}))
			defer srv.Close()

			err := SendTextToFeiShu(t.Context(), srv.URL, "secret", "hello")
			So(err, ShouldBeNil)
			So(gotReq.MsgType, ShouldEqual, "text")
			So(gotReq.Content.Text, ShouldEqual, "hello")
		})

		Convey("响应 code!=0 返回 error", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"code":19001,"msg":"sign error"}`))
			}))
			defer srv.Close()

			err := SendTextToFeiShu(t.Context(), srv.URL, "secret", "hello")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "code=19001")
		})

		Convey("HTTP 非 200 返回 error", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer srv.Close()

			err := SendTextToFeiShu(t.Context(), srv.URL, "secret", "hello")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "status=502")
		})
	})
}
