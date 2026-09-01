package htime

import (
	"strconv"
	"time"
)

// 常见时间格式常量
const (
	// TimeFormatDateTimeStandard 标准时间格式
	TimeFormatDateTimeStandard = "2006-01-02 15:04:05"
	// TimeFormatDateTimeMs 毫秒时间格式
	TimeFormatDateTimeMs = "2006-01-02 15:04:05.000"
	// TimeFormatDate 日期格式
	TimeFormatDate = "2006-01-02"
	// TimeFormatHourMinSec 时分秒格式
	TimeFormatHourMinSec = "15:04:05"
)

// TimeStampToTime 时间戳转化成time.Time类型
func TimeStampToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// TimeStampMsToTime 时间戳转化成time.Time类型
func TimeStampMsToTime(ts int64) time.Time {
	return time.UnixMilli(ts)
}

// ParseTimeStamp 解析字符串时间戳,自动判断秒/毫秒
//
// 判定阈值:数值 < 1e10 视为秒级,>= 1e10 视为毫秒级
// (1e10 秒约 2286 年,实际业务场景几乎不会越过)
func ParseTimeStamp(timeStampStr string) (time.Time, error) {
	ts, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	const secMilliBoundary = 1e10
	if ts < secMilliBoundary {
		return time.Unix(ts, 0), nil
	}
	return time.UnixMilli(ts), nil
}

func ParseTimeDataStandard(timeStr string) (time.Time, error) {
	return time.Parse(TimeFormatDateTimeStandard, timeStr)
}
