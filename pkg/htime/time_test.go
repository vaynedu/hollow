package htime

import (
	"testing"
	"time"
)

func TestTimeStampToTime(t *testing.T) {
	ts := int64(1672531200)
	expected := time.Unix(ts, 0)
	result := TimeStampToTime(ts)
	if result != expected {
		t.Errorf("TimeStampToTime(%d) = %v; want %v", ts, result, expected)
	}
}

func TestTimeStampMsToTime(t *testing.T) {
	ts := int64(1672531200000)
	expected := time.UnixMilli(ts)
	result := TimeStampMsToTime(ts)
	if result != expected {
		t.Errorf("TimeStampMsToTime(%d) = %v; want %v", ts, result, expected)
	}
}

func TestParseTimeStamp(t *testing.T) {
	// 秒级时间戳测试
	t.Run("SecondTimestamp", func(t *testing.T) {
		tsStr := "1672531200"
		expected := time.Unix(1672531200, 0)
		result, err := ParseTimeStamp(tsStr)
		if err != nil {
			t.Errorf("ParseTimeStamp(%s) returned error: %v", tsStr, err)
		}
		if result != expected {
			t.Errorf("ParseTimeStamp(%s) = %v; want %v", tsStr, result, expected)
		}
	})

	// 毫秒级时间戳测试
	t.Run("MillisecondTimestamp", func(t *testing.T) {
		tsStr := "1672531200000"
		expected := time.UnixMilli(1672531200000)
		result, err := ParseTimeStamp(tsStr)
		if err != nil {
			t.Errorf("ParseTimeStamp(%s) returned error: %v", tsStr, err)
		}
		if result != expected {
			t.Errorf("ParseTimeStamp(%s) = %v; want %v", tsStr, result, expected)
		}
	})

	// 无效时间戳测试
	t.Run("InvalidTimestamp", func(t *testing.T) {
		tsStr := "12345"
		_, err := ParseTimeStamp(tsStr)
		if err == nil {
			t.Errorf("ParseTimeStamp(%s) did not return an error for invalid timestamp", tsStr)
		}
	})
}

func TestParseTimeDataStandard(t *testing.T) {
	testCases := []struct {
		name        string
		timeStr     string
		shouldError bool
	}{
		{
			name:        "ValidStandardTime",
			timeStr:     "2023-01-01 12:00:00",
			shouldError: false,
		},
		{
			name:        "ValidStandardTimeAtMidnight",
			timeStr:     "2023-01-01 00:00:00",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected, err := time.Parse(TimeFormatDateTimeStandard, tc.timeStr)
			if err != nil && !tc.shouldError {
				t.Fatalf("Unexpected error parsing expected time: %v", err)
			}

			result, err := ParseTimeDataStandard(tc.timeStr)
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.timeStr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimeDataStandard(%s) returned error: %v", tc.timeStr, err)
			}
			if !result.Equal(expected) {
				t.Errorf("ParseTimeDataStandard(%s) = %v; want %v", tc.timeStr, result, expected)
			}
		})
	}
}
