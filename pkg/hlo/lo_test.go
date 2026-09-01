package hlo

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoSamples(t *testing.T) {
	Convey("lo.Map", t, func() {
		So(lo.Map([]int{1, 2, 3, 4, 5}, func(item, _ int) int { return item * 2 }),
			ShouldResemble, []int{2, 4, 6, 8, 10})
	})

	Convey("lo.Filter", t, func() {
		So(lo.Filter([]int{1, 2, 3, 4, 5}, func(item, _ int) bool { return item%2 == 0 }),
			ShouldResemble, []int{2, 4})
	})

	Convey("lo.Reduce / ReduceRight", t, func() {
		sum := func(agg, item, _ int) int { return agg + item }
		So(lo.Reduce([]int{1, 2, 3, 4, 5}, sum, 0), ShouldEqual, 15)
		So(lo.ReduceRight([]int{1, 2, 3, 4, 5}, sum, 0), ShouldEqual, 15)
	})

	Convey("lo.SliceToMap", t, func() {
		got := lo.SliceToMap([]int{1, 2, 3}, func(v int) (string, int) {
			return fmt.Sprintf("key-%d", v), v
		})
		So(got, ShouldResemble, map[string]int{"key-1": 1, "key-2": 2, "key-3": 3})
	})
}
