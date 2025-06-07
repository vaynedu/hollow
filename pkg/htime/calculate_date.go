package htime

import (
	"errors"
	"time"
)

// GetTotalDaysInMonth 获取指定年月的总天数
func GetTotalDaysInMonth(year, month int) int {
	// 获取指定年月的下个月第一天
	nextMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	// 下个月第一天减去一天即为该月最后一天
	lastDay := nextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}

// GetCurrentYearMonthDay 获取当前时间年、月、日
func GetCurrentYearMonthDay() (int, int, int) {
	now := time.Now()
	return now.Year(), int(now.Month()), now.Day()
}

// GetCurrentTimeString 获取当前时间字符串，年月日时分秒
func GetCurrentTimeString() string {
	return time.Now().Format(TimeFormatDateTimeStandard)
}

// GetSecondsSinceMidnight 获取当天0点到当前时间的秒数
func GetSecondsSinceMidnight() int64 {
	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return int64(now.Sub(todayMidnight).Seconds())
}

// GetSecondsUntilMidnight 获取当前到当天23:59:59的秒数
func GetSecondsUntilMidnight() int64 {
	now := time.Now()
	tomorrowMidnight := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return int64(tomorrowMidnight.Sub(now).Seconds())
}

// GetDayOfMonth 获取某一年某个月的特定一天
func GetDayOfMonth(year, month, day int) (time.Time, error) {
	if month < 1 || month > 12 || day < 1 {
		return time.Time{}, errors.New("invalid month or day")
	}
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
	if day > daysInMonth {
		return time.Time{}, errors.New("day out of range for month")
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// IsValidDate 检查年月日是否有效
func IsValidDate(year, month, day int) bool {
	if year < 0 || month < 1 || month > 12 || day < 1 {
		return false
	}
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
	if day > daysInMonth {
		return false
	}
	return true
}
