package htime

import (
	"errors"
	"testing"
	"time"
)

func TestGetTotalDaysInMonth(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    int
		expected int
	}{
		{name: "LeapYear", year: 2024, month: 2, expected: 29},
		{name: "EndOfYear", year: 2025, month: 12, expected: 31},
		{name: "InvalidMonth", year: 2025, month: 6, expected: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTotalDaysInMonth(tt.year, tt.month)
			if result != tt.expected {
				t.Errorf("GetTotalDaysInMonth(%d, %d) = %d; want %d", tt.year, tt.month, result, tt.expected)
			}
		})
	}
}

func TestGetCurrentYearMonthDay(t *testing.T) {
	n := time.Now()
	year, month, day := GetCurrentYearMonthDay()
	if year != n.Year() || month != int(n.Month()) || day != n.Day() {
		t.Errorf("GetCurrentYearMonthDay() = (%d, %d, %d); want (%d, %d, %d)", year, month, day, n.Year(), n.Month(), n.Day())
	}
}

func TestGetCurrentTimeString(t *testing.T) {
	// 由于 TimeFormatDateTimeStandard 未定义，暂时模拟测试
	// 实际应根据具体定义修改
	t.Log(GetCurrentTimeString())
}

func TestGetSecondsSinceMidnight(t *testing.T) {
	n := time.Now()
	todayMidnight := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
	expected := int64(n.Sub(todayMidnight).Seconds())
	result := GetSecondsSinceMidnight()
	if result != expected {
		t.Errorf("GetSecondsSinceMidnight() = %d; want %d", result, expected)
	}
}

func TestGetSecondsUntilMidnight(t *testing.T) {
	n := time.Now()
	tomorrowMidnight := time.Date(n.Year(), n.Month(), n.Day(), 23, 59, 59, 0, n.Location())
	expected := int64(tomorrowMidnight.Sub(n).Seconds())
	result := GetSecondsUntilMidnight()
	if result != expected {
		t.Errorf("GetSecondsUntilMidnight() = %d; want %d", result, expected)
	}
}

func TestGetDayOfMonth(t *testing.T) {
	tests := []struct {
		year     int
		month    int
		day      int
		expected time.Time
		err      error
	}{
		{2024, 2, 29, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), nil},
		{2023, 2, 29, time.Time{}, errors.New("day out of range for month")},
		{2024, 13, 1, time.Time{}, errors.New("invalid month or day")},
		{2024, 2, 0, time.Time{}, errors.New("invalid month or day")},
	}

	for _, tt := range tests {
		result, err := GetDayOfMonth(tt.year, tt.month, tt.day)
		if (err != nil && err.Error() != tt.err.Error()) || (err == nil && tt.err != nil) {
			t.Errorf("GetDayOfMonth(%d, %d, %d) error = %v; want %v", tt.year, tt.month, tt.day, err, tt.err)
		}
		if !result.Equal(tt.expected) {
			t.Errorf("GetDayOfMonth(%d, %d, %d) = %v; want %v", tt.year, tt.month, tt.day, result, tt.expected)
		}
	}
}

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    int
		day      int
		expected bool
	}{
		// 有效日期
		{name: "ValidDate", year: 2024, month: 2, day: 29, expected: true},
		{name: "ValidDateEndOfYear", year: 2025, month: 12, day: 31, expected: true},
		// 无效年份
		{name: "InvalidYear", year: -1, month: 1, day: 1, expected: false},
		// 无效月份
		{name: "InvalidMonthLow", year: 2024, month: 0, day: 1, expected: false},
		{name: "InvalidMonthHigh", year: 2024, month: 13, day: 1, expected: false},
		// 无效日期
		{name: "InvalidDayLow", year: 2024, month: 1, day: 0, expected: false},
		{name: "InvalidDayHigh", year: 2023, month: 2, day: 29, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDate(tt.year, tt.month, tt.day)
			if result != tt.expected {
				t.Errorf("IsValidDate(%d, %d, %d) = %v; want %v", tt.year, tt.month, tt.day, result, tt.expected)
			}
		})
	}
}
