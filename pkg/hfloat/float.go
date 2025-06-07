package hfloat

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"strconv"
)

// float 比较的最小值
const floatCompareMin = 0.000001

// IsFloatEqual 两个 float 类型的值是否相等
// 由于浮点数存在精度问题，不能直接使用 == 比较，而是通过比较两个数的差值是否小于一个极小值来判断
func IsFloatEqual(a, b float64) bool {
	return abs(a-b) < floatCompareMin
}

// abs 计算浮点数的绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// CompareFloat 比较两个 float 类型的值大小
// 返回值：
// -1 表示 a 小于 b
//
//	0 表示 a 等于 b
//	1 表示 a 大于 b
func CompareFloat(a, b float64) int {
	diff := a - b
	if abs(diff) < floatCompareMin {
		return 0
	} else if diff < 0 {
		return -1
	}
	return 1
}

// RoundUpFloat 向上取整函数
func RoundUpFloat(x float64) float64 {
	return math.Round(x*100) / 100
}

// RoundDownFloat 向下取整函数
func RoundDownFloat(x float64) float64 {
	return math.Floor(x*100) / 100
}

// ConvertFloatToString converts a float value to a string
func ConvertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ConvertStringToFloat 将字符串转换为 float 类型的值
func ConvertStringToFloat(s string) (float64, error) {
	ss, err := decimal.NewFromString(s)
	if err != nil {
		return 0, err
	}
	res, _ := ss.Float64()
	return res, nil
}

// AddFloat 两个 float 类型的值相加
func AddFloat(a, b float64) float64 {
	aa := decimal.NewFromFloat(a)
	bb := decimal.NewFromFloat(b)
	res := aa.Add(bb)
	f, _ := res.Float64()
	return f
}

// SubtractFloat 两个 float 类型的值相减
func SubtractFloat(a, b float64) float64 {
	aa := decimal.NewFromFloat(a)
	bb := decimal.NewFromFloat(b)
	res := aa.Sub(bb)
	f, _ := res.Float64()
	return f
}

// MultiplyFloat 两个 float 类型的值相乘
func MultiplyFloat(a, b float64) float64 {
	aa := decimal.NewFromFloat(a)
	bb := decimal.NewFromFloat(b)
	res := aa.Mul(bb)
	f, _ := res.Float64()
	return f
}

// DivideFloat 两个 float 类型的值相除
func DivideFloat(a, b float64) (float64, error) {
	if abs(b) < floatCompareMin {
		return 0.0, fmt.Errorf("cannot divide by zero")
	}

	aa := decimal.NewFromFloat(a)
	bb := decimal.NewFromFloat(b)
	res := aa.Div(bb)
	f, _ := res.Float64()
	return f, nil
}

// AddStringFloat 两个 string 类型 float 相加
func AddStringFloat(a, b string) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "", err
	}
	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "", err
	}
	res := aa.Add(bb)
	return res.String(), nil
}
