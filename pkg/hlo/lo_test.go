package hlo

// lo 库 使用
// github.com/samber/lo
import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	// 测试 Map 函数
	slice := []int{1, 2, 3, 4, 5}
	mapped := lo.Map(slice, func(item int, index int) int {
		return item * 2
	})
	assert.Equal(t, []int{2, 4, 6, 8, 10}, mapped)
}

func TestFilter(t *testing.T) {
	// 测试 Filter 函数
	slice := []int{1, 2, 3, 4, 5}
	filtered := lo.Filter(slice, func(item int, index int) bool {
		return item%2 == 0
	})
	assert.Equal(t, []int{2, 4}, filtered)
}

func TestReduce(t *testing.T) {
	// 测试 Reduce 函数
	slice := []int{1, 2, 3, 4, 5}
	reduced := lo.Reduce(slice, func(agg int, item int, index int) int {
		return agg + item
	}, 0)
	assert.Equal(t, 15, reduced)
}

func TestReduceRight(t *testing.T) {
	// 测试 ReduceRight 函数
	slice := []int{1, 2, 3, 4, 5}
	reduced := lo.ReduceRight(slice, func(agg int, item int, index int) int {
		return agg + item
	}, 0)
	assert.Equal(t, 15, reduced)
}

func TestSliceToMap(t *testing.T) {
	// 测试 SliceToMap 函数
	slice := []int{1, 2, 3, 4, 5}
	mapped := lo.SliceToMap(slice, func(l int) (string, int) {
		return fmt.Sprintf("key-%d", l), l
	})
	assert.Equal(t, map[string]int{"key-1": 1, "key-2": 2, "key-3": 3, "key-4": 4, "key-5": 5}, mapped)
}
