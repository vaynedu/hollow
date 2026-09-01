package hfloat

import (
	"fmt"
	"math"
	"strconv"

	"github.com/shopspring/decimal"
)

const floatCompareMin = 0.000001

// IsFloatEqual 判断两个 float64 是否在精度容差内相等
func IsFloatEqual(a, b float64) bool {
	return math.Abs(a-b) < floatCompareMin
}

// CompareFloat 比较两个 float64
//
//	-1: a < b
//	 0: a == b(在容差内)
//	 1: a > b
func CompareFloat(a, b float64) int {
	diff := a - b
	if math.Abs(diff) < floatCompareMin {
		return 0
	} else if diff < 0 {
		return -1
	}
	return 1
}

// RoundUpFloat 四舍五入保留 2 位小数
func RoundUpFloat(x float64) float64 {
	return math.Round(x*100) / 100
}

// RoundDownFloat 向下取整保留 2 位小数
func RoundDownFloat(x float64) float64 {
	return math.Floor(x*100) / 100
}

// ConvertFloatToString float64 转字符串(保留有效精度)
func ConvertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ConvertStringToFloat 字符串转 float64,走 decimal 避免精度问题
func ConvertStringToFloat(s string) (float64, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return 0, err
	}
	res, _ := d.Float64()
	return res, nil
}

// AddFloat 用 decimal 做加法避免精度丢失
func AddFloat(a, b float64) float64 {
	res := decimal.NewFromFloat(a).Add(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f
}

// SubtractFloat 用 decimal 做减法
func SubtractFloat(a, b float64) float64 {
	res := decimal.NewFromFloat(a).Sub(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f
}

// MultiplyFloat 用 decimal 做乘法
func MultiplyFloat(a, b float64) float64 {
	res := decimal.NewFromFloat(a).Mul(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f
}

// DivideFloat 用 decimal 做除法,除数为 0 返回 error
func DivideFloat(a, b float64) (float64, error) {
	if math.Abs(b) < floatCompareMin {
		return 0, fmt.Errorf("hfloat: cannot divide by zero")
	}
	res := decimal.NewFromFloat(a).Div(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f, nil
}

// AddStringFloat 两个字符串形式的小数相加,返回字符串结果
func AddStringFloat(a, b string) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "", err
	}
	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "", err
	}
	return aa.Add(bb).String(), nil
}
