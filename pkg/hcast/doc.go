// Package hcast 是 github.com/spf13/cast 的"推荐入口"包装。
//
// 本身不提供任何函数。业务代码可以 import 本包以声明对类型转换工具的依赖,
// 以便未来如果 hollow 切换到自有实现时,只在本包内做透明替换。
//
// 当前直接使用 spf13/cast 即可:
//
//	import "github.com/spf13/cast"
//
//	i := cast.ToInt64("123")
//	s := cast.ToString(time.Now())
package hcast
