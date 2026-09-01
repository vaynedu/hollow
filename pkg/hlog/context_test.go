package hlog

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

func TestContextLogger(t *testing.T) {
	Convey("Context Logger", t, func() {
		original := L()
		Reset(func() { SetDefault(original) })

		Convey("默认 Logger 永不为 nil", func() {
			SetDefault(nil)
			So(L(), ShouldNotBeNil)
			So(FromContext(context.Background()), ShouldEqual, L())
		})

		Convey("返回 Context 中注入的 Logger", func() {
			logger := zap.NewNop().Named("request")
			ctx := NewContext(context.Background(), logger)

			So(FromContext(ctx), ShouldEqual, logger)
		})

		Convey("nil Context 和 nil Logger 安全回退", func() {
			SetDefault(zap.NewNop().Named("default"))
			ctx := NewContext(nil, nil)

			So(FromContext(nil), ShouldEqual, L())
			So(FromContext(ctx), ShouldEqual, L())
		})
	})
}
