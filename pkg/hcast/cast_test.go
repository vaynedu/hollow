package hcast

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cast"
)

func TestCast(t *testing.T) {
	Convey("spf13/cast 基础用法验证", t, func() {
		So(cast.ToString(1.23456789), ShouldEqual, "1.23456789")
		So(cast.ToString(123456789), ShouldEqual, "123456789")
		So(cast.ToString(nil), ShouldEqual, "")
		So(cast.ToInt64("12344"), ShouldEqual, int64(12344))
		So(cast.ToFloat64("12344"), ShouldEqual, float64(12344))
	})
}

func TestCastTime(t *testing.T) {
	Convey("cast.ToTime/ToString time.Time", t, func() {
		now := time.Now()
		So(cast.ToTime(now).Equal(now), ShouldBeTrue)
		So(cast.ToString(now), ShouldNotBeEmpty)
	})
}
