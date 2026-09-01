// Package hlo 是 github.com/samber/lo 的"推荐入口"包装。
//
// 本身不提供任何函数。业务代码可以 import 本包以声明对函数式工具的依赖,
// 以便未来 hollow 切换到自有实现时透明替换。
//
// 当前直接使用 samber/lo 即可:
//
//	import "github.com/samber/lo"
//
//	doubled := lo.Map([]int{1, 2, 3}, func(v, _ int) int { return v * 2 })
//	evens := lo.Filter([]int{1, 2, 3, 4}, func(v, _ int) bool { return v%2 == 0 })
package hlo
