package hes

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	Convey("NewClient", t, func() {
		Convey("Addresses 与 CloudID 都为空返回 ErrEmptyAddresses", func() {
			_, err := NewClient(context.Background(), Config{})
			So(err, ShouldEqual, ErrEmptyAddresses)
		})

		Convey("无效 Addresses Ping 失败返错误", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			_, err := NewClient(ctx, Config{
				Addresses: []string{"http://127.0.0.1:1"}, // 几乎一定连不上
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "ping failed")
		})
	})
}

func TestPing(t *testing.T) {
	Convey("Ping", t, func() {
		Convey("nil client 返回错误", func() {
			err := Ping(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "nil client")
		})
	})
}
