package hredlock

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// 初始化测试环境
	// 执行测试用例
	exitCode := m.Run()

	// 清理测试环境
	// 退出测试
	os.Exit(exitCode)
}

func Test_Redsync(t *testing.T) {
	// redsync开源包
	// https://juejin.cn/post/7241400783676424247?searchId=20240626142724F4046686ABCFD33C5163#heading-2
	// https://juejin.cn/post/7384750303521292303
	// https://github.com/go-redsync/redsync
	t.Log("redsync测试")
}
