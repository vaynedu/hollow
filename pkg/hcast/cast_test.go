package hcast

import (
	"fmt"
	"github.com/spf13/cast"
	"testing"
	"time"
)

func TestCast(t *testing.T) {
	fmt.Println(cast.ToString(1.23456789))
	fmt.Println(cast.ToString(123456789))
	fmt.Println(cast.ToString(nil))

	fmt.Println(cast.ToInt64("12344"))
	fmt.Println(cast.ToFloat64("12344"))
}

func TestCastTime(t *testing.T) {
	cast.ToTime(time.Now())
	fmt.Println(cast.ToString(time.Now()))
}
