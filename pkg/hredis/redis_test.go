package hredis

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	Convey("NewClient", t, func() {
		Convey("空 Addr 返回 ErrEmptyAddr", func() {
			_, err := NewClient(context.Background(), Config{})
			So(err, ShouldEqual, ErrEmptyAddr)
		})

		Convey("无效 Addr Ping 失败返回错误", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			_, err := NewClient(ctx, Config{
				Addr:        "127.0.0.1:1", // 几乎一定连不上
				DialTimeout: 100 * time.Millisecond,
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
