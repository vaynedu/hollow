package htime

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTotalDaysInMonth(t *testing.T) {
	Convey("GetTotalDaysInMonth", t, func() {
		So(GetTotalDaysInMonth(2024, 2), ShouldEqual, 29)
		So(GetTotalDaysInMonth(2025, 12), ShouldEqual, 31)
		So(GetTotalDaysInMonth(2025, 6), ShouldEqual, 30)
	})
}

func TestGetCurrentYearMonthDay(t *testing.T) {
	Convey("GetCurrentYearMonthDay 与 time.Now 一致", t, func() {
		n := time.Now()
		y, m, d := GetCurrentYearMonthDay()
		So(y, ShouldEqual, n.Year())
		So(m, ShouldEqual, int(n.Month()))
		So(d, ShouldEqual, n.Day())
	})
}

func TestGetCurrentTimeString(t *testing.T) {
	Convey("GetCurrentTimeString 长度等于标准格式", t, func() {
		s := GetCurrentTimeString()
		So(len(s), ShouldEqual, len(TimeFormatDateTimeStandard))
	})
}

func TestGetSecondsSinceMidnight(t *testing.T) {
	Convey("GetSecondsSinceMidnight 与本地时间一致", t, func() {
		n := time.Now()
		mid := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
		So(GetSecondsSinceMidnight(), ShouldEqual, int64(n.Sub(mid).Seconds()))
	})
}

func TestGetSecondsUntilMidnight(t *testing.T) {
	Convey("GetSecondsUntilMidnight 与本地时间一致", t, func() {
		n := time.Now()
		end := time.Date(n.Year(), n.Month(), n.Day(), 23, 59, 59, 0, n.Location())
		So(GetSecondsUntilMidnight(), ShouldEqual, int64(end.Sub(n).Seconds()))
	})
}

func TestGetDayOfMonth(t *testing.T) {
	Convey("GetDayOfMonth", t, func() {
		got, err := GetDayOfMonth(2024, 2, 29)
		So(err, ShouldBeNil)
		So(got.Equal(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)), ShouldBeTrue)

		_, err = GetDayOfMonth(2023, 2, 29)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "day out of range for month")

		_, err = GetDayOfMonth(2024, 13, 1)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid month or day")

		_, err = GetDayOfMonth(2024, 2, 0)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid month or day")
	})
}

func TestIsValidDate(t *testing.T) {
	Convey("IsValidDate", t, func() {
		So(IsValidDate(2024, 2, 29), ShouldBeTrue)
		So(IsValidDate(2025, 12, 31), ShouldBeTrue)
		So(IsValidDate(-1, 1, 1), ShouldBeFalse)
		So(IsValidDate(2024, 0, 1), ShouldBeFalse)
		So(IsValidDate(2024, 13, 1), ShouldBeFalse)
		So(IsValidDate(2024, 1, 0), ShouldBeFalse)
		So(IsValidDate(2023, 2, 29), ShouldBeFalse)
	})
}
