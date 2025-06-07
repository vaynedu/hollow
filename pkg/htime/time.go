package htime

import (
	"errors"
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

func ParseTimeStamp(timeStampStr string) (time.Time, error) {
	timeStamp, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	var res time.Time
	if len(timeStampStr) == 10 { // 秒级
		res = time.Unix(timeStamp, 0)
	} else if len(timeStampStr) == 13 { // 毫米级
		res = time.UnixMilli(timeStamp)
	} else {
		return time.Time{}, errors.New("invalid timestamp")
	}
	return res, nil
}

func ParseTimeDataStandard(timeStr string) (time.Time, error) {
	return time.Parse(TimeFormatDateTimeStandard, timeStr)
}
