package hfloat

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsFloatEqual(t *testing.T) {
	Convey("IsFloatEqual", t, func() {
		So(IsFloatEqual(1.0, 1.0), ShouldBeTrue)
		So(IsFloatEqual(1.0, 1.0000001), ShouldBeTrue)
		So(IsFloatEqual(1.0, 1.00001), ShouldBeFalse)
		// 容差为 1e-6,差值 1e-7 在容差内
		So(IsFloatEqual(0.0, -0.0000001), ShouldBeTrue)
		So(IsFloatEqual(0.0, 1.0), ShouldBeFalse)
	})
}

func TestCompareFloat(t *testing.T) {
	Convey("CompareFloat", t, func() {
		So(CompareFloat(1.0, 2.0), ShouldEqual, -1)
		So(CompareFloat(2.0, 1.0), ShouldEqual, 1)
		So(CompareFloat(1.0, 1.0000001), ShouldEqual, 0)
	})
}

func TestRoundUpFloat(t *testing.T) {
	Convey("RoundUpFloat 保留 2 位", t, func() {
		So(RoundUpFloat(1.234), ShouldEqual, 1.23)
		So(RoundUpFloat(1.235), ShouldEqual, 1.24)
	})
}

func TestRoundDownFloat(t *testing.T) {
	Convey("RoundDownFloat 向下取 2 位", t, func() {
		So(RoundDownFloat(1.234), ShouldEqual, 1.23)
		So(RoundDownFloat(1.239), ShouldEqual, 1.23)
	})
}

func TestConvertFloatToString(t *testing.T) {
	Convey("ConvertFloatToString", t, func() {
		cases := []struct {
			in  float64
			out string
		}{
			{1.23, "1.23"},
			{1.0, "1"},
			{0.0, "0"},
			{0.000001, "0.000001"},
			{-1.23, "-1.23"},
		}
		for _, c := range cases {
			So(ConvertFloatToString(c.in), ShouldEqual, c.out)
		}
	})
}

func TestConvertStringToFloat(t *testing.T) {
	Convey("ConvertStringToFloat", t, func() {
		Convey("合法数值", func() {
			v, err := ConvertStringToFloat("1.23")
			So(err, ShouldBeNil)
			So(IsFloatEqual(v, 1.23), ShouldBeTrue)
		})
		Convey("非法字符串返回 error", func() {
			_, err := ConvertStringToFloat("abc")
			So(err, ShouldNotBeNil)
		})
		Convey("空字符串返回 error", func() {
			_, err := ConvertStringToFloat("")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestArithmetic(t *testing.T) {
	Convey("加减乘除", t, func() {
		So(IsFloatEqual(AddFloat(1.23, 4.56), 5.79), ShouldBeTrue)
		So(IsFloatEqual(SubtractFloat(4.56, 1.23), 3.33), ShouldBeTrue)
		So(IsFloatEqual(MultiplyFloat(2.0, 3.0), 6.0), ShouldBeTrue)

		v, err := DivideFloat(6.0, 2.0)
		So(err, ShouldBeNil)
		So(IsFloatEqual(v, 3.0), ShouldBeTrue)

		_, err = DivideFloat(1.0, 0.0)
		So(err, ShouldNotBeNil)
	})
}

func TestAddStringFloat(t *testing.T) {
	Convey("AddStringFloat", t, func() {
		v, err := AddStringFloat("1.23", "4.56")
		So(err, ShouldBeNil)
		So(v, ShouldEqual, "5.79")

		_, err = AddStringFloat("abc", "1.0")
		So(err, ShouldNotBeNil)
	})
}
