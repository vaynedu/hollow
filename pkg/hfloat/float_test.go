package hfloat

import (
	"testing"
)

func TestIsFloatEqual(t *testing.T) {
	tests := []struct {
		a        float64
		b        float64
		expected bool
	}{
		{1.0, 1.0, true},
		{1.0, 1.0000001, true},
		{1.0, 1.00001, false},
		{0.0, 0.0, true},
		{0.0, 1.0, false},
		{0.0, -0.00001, true},
	}

	for _, tt := range tests {
		result := IsFloatEqual(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("IsFloatEqual(%f, %f) = %v; want %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestCompareFloat(t *testing.T) {
	tests := []struct {
		a        float64
		b        float64
		expected int
	}{
		{1.0, 2.0, -1},
		{2.0, 1.0, 1},
		{1.0, 1.0000001, 0},
	}

	for _, tt := range tests {
		result := CompareFloat(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("CompareFloat(%f, %f) = %d; want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestRoundUpFloat(t *testing.T) {
	tests := []struct {
		x        float64
		expected float64
	}{
		{1.234, 1.23},
		{1.235, 1.24},
	}

	for _, tt := range tests {
		result := RoundUpFloat(tt.x)
		if result != tt.expected {
			t.Errorf("RoundUpFloat(%f) = %f; want %f", tt.x, result, tt.expected)
		}
	}
}

func TestRoundDownFloat(t *testing.T) {
	tests := []struct {
		x        float64
		expected float64
	}{
		{1.234, 1.23},
		{1.239, 1.23},
	}

	for _, tt := range tests {
		result := RoundDownFloat(tt.x)
		if result != tt.expected {
			t.Errorf("RoundDownFloat(%f) = %f; want %f", tt.x, result, tt.expected)
		}
	}
}

func TestConvertFloatToString(t *testing.T) {
	tests := []struct {
		f        float64
		expected string
	}{
		{1.23, "1.23"},
		{1.0, "1"},
		{0.0, "0"},
		{0.000001, "0.000001"},
		{123456789.123456789, "123456789.12345679"}, // 注意精度损失 最多保留 8 位小数
		{1234567890.123456789, "1234567890.1234567"},
		{12345678900.123456789, "12345678900.123457"},
		{123456789000.123456789, "123456789000.12346"},

		{-1.23, "-1.23"},
	}

	for _, tt := range tests {
		result := ConvertFloatToString(tt.f)
		if result != tt.expected {
			t.Errorf("ConvertFloatToString(%f) = %s; want %s", tt.f, result, tt.expected)
		}
	}
}

func TestConvertStringToFloat(t *testing.T) {
	tests := []struct {
		s        string
		expected float64
		error    bool
	}{
		{"1.23", 1.23, false},
		{"abc", 0, true},
		{"123", 123, false},
		{"123.456", 123.456, false},
		{"-1.23", -1.23, false},
		{"", 0, true},
	}

	for _, tt := range tests {
		result, err := ConvertStringToFloat(tt.s)
		if (err != nil) != tt.error {
			t.Errorf("ConvertStringToFloat(%s) error = %v; want error %v", tt.s, err, tt.error)
		}
		if result != tt.expected {
			t.Errorf("ConvertStringToFloat(%s) = %f; want %f", tt.s, result, tt.expected)
		}
	}
}

func TestAddFloat(t *testing.T) {
	tests := []struct {
		a        float64
		b        float64
		expected float64
	}{
		{1.23, 4.56, 5.79},
		{-1.0, 1.0, 0.0},
		{0.0, 0.0, 0.0},
		{0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		result := AddFloat(tt.a, tt.b)
		if !IsFloatEqual(result, tt.expected) {
			t.Errorf("AddFloat(%f, %f) = %f; want %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestSubtractFloat(t *testing.T) {
	tests := []struct {
		a        float64
		b        float64
		expected float64
	}{
		{4.56, 1.23, 3.33},
		{1.0, -1.0, 2.0},
	}

	for _, tt := range tests {
		result := SubtractFloat(tt.a, tt.b)
		if !IsFloatEqual(result, tt.expected) {
			t.Errorf("SubtractFloat(%f, %f) = %f; want %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMultiplyFloat(t *testing.T) {
	tests := []struct {
		a        float64
		b        float64
		expected float64
	}{
		{2.0, 3.0, 6.0},
		{0.0, 5.0, 0.0},
	}

	for _, tt := range tests {
		result := MultiplyFloat(tt.a, tt.b)
		if !IsFloatEqual(result, tt.expected) {
			t.Errorf("MultiplyFloat(%f, %f) = %f; want %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestDivideFloat(t *testing.T) {
	tests := []struct {
		a        float64
		b        float64
		expected float64
		error    bool
	}{
		{6.0, 2.0, 3.0, false},
		{5.0, 0.0, 0.0, true},
		{0.0, 0, 0.0, true},
	}

	for _, tt := range tests {
		result, err := DivideFloat(tt.a, tt.b)
		if (err != nil) != tt.error {
			t.Errorf("DivideFloat(%f, %f) error = %v; want error %v", tt.a, tt.b, err, tt.error)
		}
		if !IsFloatEqual(result, tt.expected) {
			t.Errorf("DivideFloat(%f, %f) = %f; want %f", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestAddStringFloat(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected string
		error    bool
	}{
		{"1.23", "4.56", "5.79", false},
		{"abc", "1.0", "", true},
		{"123", "456", "579", false},
		{"-1.23", "1.23", "0", false},
		{"", "", "", true},
	}

	for _, tt := range tests {
		result, err := AddStringFloat(tt.a, tt.b)
		if (err != nil) != tt.error {
			t.Errorf("AddStringFloat(%s, %s) error = %v; want error %v", tt.a, tt.b, err, tt.error)
		}
		if result != tt.expected {
			t.Errorf("AddStringFloat(%s, %s) = %s; want %s", tt.a, tt.b, result, tt.expected)
		}
	}
}
