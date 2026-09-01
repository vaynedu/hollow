package htime

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTimeStampToTime(t *testing.T) {
	Convey("TimeStampToTime", t, func() {
		ts := int64(1672531200)
		So(TimeStampToTime(ts).Equal(time.Unix(ts, 0)), ShouldBeTrue)
	})
}

func TestTimeStampMsToTime(t *testing.T) {
	Convey("TimeStampMsToTime", t, func() {
		ts := int64(1672531200000)
		So(TimeStampMsToTime(ts).Equal(time.UnixMilli(ts)), ShouldBeTrue)
	})
}

func TestParseTimeStamp(t *testing.T) {
	Convey("ParseTimeStamp", t, func() {
		Convey("秒级 10 位", func() {
			got, err := ParseTimeStamp("1672531200")
			So(err, ShouldBeNil)
			So(got.Equal(time.Unix(1672531200, 0)), ShouldBeTrue)
		})
		Convey("毫秒级 13 位", func() {
			got, err := ParseTimeStamp("1672531200000")
			So(err, ShouldBeNil)
			So(got.Equal(time.UnixMilli(1672531200000)), ShouldBeTrue)
		})
		Convey("小数值视为秒级", func() {
			got, err := ParseTimeStamp("12345")
			So(err, ShouldBeNil)
			So(got.Equal(time.Unix(12345, 0)), ShouldBeTrue)
		})
		Convey("非数字字符串返回 error", func() {
			_, err := ParseTimeStamp("not-a-number")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestParseTimeDataStandard(t *testing.T) {
	Convey("ParseTimeDataStandard", t, func() {
		got, err := ParseTimeDataStandard("2023-01-01 12:00:00")
		So(err, ShouldBeNil)
		expected, _ := time.Parse(TimeFormatDateTimeStandard, "2023-01-01 12:00:00")
		So(got.Equal(expected), ShouldBeTrue)
	})
}
