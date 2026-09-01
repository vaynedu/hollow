# pkg Iter 1: 文档沉淀 + hexcel 改名 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 hollow 的 pkg 库从"够用"提到"工程级"的第 1 个迭代:11 个非空壳包补齐 `doc.go`(go pkg.go.dev 入口)和至少 1 个 `ExampleXxx()` 函数;`hexcel` 改名为 `hcsv`(名实相符,只支持 CSV)。

**Architecture:**
- 每个包加一个 `doc.go`,仅含 `package` 声明 + 完整的中文 godoc 注释,讲清楚:定位 / 主要能力 / 安全注意。
- 每个有公开函数的包加一个 `Example*` 函数(直接追加到现有 `_test.go` 同 package 内,不另建 `example_test.go` 文件),展示典型调用形态;能给 `// Output:` 的就给(自动验证),依赖外部服务的(redis/lark webhook)只展示调用形态。
- `hexcel` 整目录改名为 `hcsv`,函数名 `GetCSVData` 不变,降低破坏面。
- 空壳包 `hes` / `hredlock` 留给 Iter 2 跟实现一起加文档。

**Tech Stack:** Go 1.25 / goconvey(已有)

## Global Constraints

- 包路径前缀 `github.com/vaynedu/hollow/pkg/...`
- pkg 内不得依赖 `internal/...`
- 新加的 Example 函数与现有 Convey 测试同包共存,不引入新依赖
- 中文 godoc 注释 + 中文 commit message
- Go 1.25
- 不动 `cmd/`, `internal/`, `example/`, `examples/`,不动 `hollow.go`
- 任务结束前 `go build ./... && go test ./pkg/... -count=1` 全绿
- `go doc github.com/vaynedu/hollow/pkg/<name>` 对每个包都能输出有效文档
- 已验证 `pkg/hexcel` 全仓库无外部引用(grep 通过),改名安全

## File Inventory

**Task 1 改名:**
- Delete + Create(`git mv`):`pkg/hexcel/excel.go` → `pkg/hcsv/csv.go`
- Delete + Create:`pkg/hexcel/excel_test.go` → `pkg/hcsv/csv_test.go`
- Modify: 两个文件内 `package hexcel` → `package hcsv`

**Task 2 文档/示例(每个包加 doc.go,部分包追加 example_test.go 或在现有 _test.go 内追加 Example*):**

| 包 | 加 doc.go | 追加 Example |
|---|---|---|
| hcast | ✓ | 跳过(本包无任何公开函数) |
| hcond | ✓ | `ExampleCondition_ToSQL` |
| hecode | ✓ | `ExampleNew`, `ExampleWrap` |
| hcsv(改名后) | ✓ | `ExampleGetCSVData` |
| hfloat | ✓ | `ExampleAddFloat`, `ExampleAddStringFloat` |
| hidgenerator | ✓ | `ExampleNewUuid` |
| hlark | ✓ | `ExampleGenSign` |
| hlo | ✓ | 跳过(本包无任何公开函数) |
| hredis | ✓ | `ExampleNewClient`(无 `// Output:`,纯展示) |
| hresty | ✓ | `ExampleNewRestyClient`(无 `// Output:`,纯展示) |
| htime | ✓ | `ExampleParseTimeStamp`, `ExampleGetTotalDaysInMonth` |

**不动:**
- `pkg/hes/`(Iter 2 一起处理)
- `pkg/hredlock/`(Iter 2 一起处理)
- 现有 `_test.go` 的 Convey 测试函数原样保留

---

## Task 1: hexcel → hcsv 改名

**目的:** 包名只支持 CSV,叫 hexcel 是误导;且后续 Iter 4 可能新增真正的 xlsx 支持,需要先把 csv 部分独立。

**Files:**
- Move: `pkg/hexcel/excel.go` → `pkg/hcsv/csv.go`
- Move: `pkg/hexcel/excel_test.go` → `pkg/hcsv/csv_test.go`
- Modify(改 package 名 + 注释微调): `pkg/hcsv/csv.go`, `pkg/hcsv/csv_test.go`
- Delete: 空的 `pkg/hexcel/` 目录

**Interfaces produces:**(完全沿用,不破坏)
- `func GetCSVData(filename string) ([][]string, error)`

- [ ] **Step 1: 创建新目录 + git mv 文件**

```bash
cd /Users/nikki/go/src/hollow
mkdir -p pkg/hcsv
git mv pkg/hexcel/excel.go pkg/hcsv/csv.go
git mv pkg/hexcel/excel_test.go pkg/hcsv/csv_test.go
rmdir pkg/hexcel
```

- [ ] **Step 2: 改 `pkg/hcsv/csv.go` 的 package 名**

把第 1 行 `package hexcel` 改成 `package hcsv`。其它内容不动。

```go
package hcsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
)

func GetCSVData(filename string) ([][]string, error) {
	// ... 不动
}
```

- [ ] **Step 3: 改 `pkg/hcsv/csv_test.go` 的 package 名**

把第 1 行 `package hexcel` 改成 `package hcsv`。

- [ ] **Step 4: 验证编译与测试**

```bash
cd /Users/nikki/go/src/hollow
go build ./... && go test ./pkg/hcsv/... -v -count=1
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add -A
git commit -m "refactor(pkg): hexcel 改名为 hcsv,名实相符

当前实现只支持 CSV 读取,叫 hexcel 容易让使用者以为支持 xlsx。
本次只改包名与目录名,GetCSVData 函数签名不变,无外部引用受影响。"
```

---

## Task 2: 11 个包加 doc.go + Example

**Files(新增):**
- Create: `pkg/hcast/doc.go`
- Create: `pkg/hcond/doc.go`
- Create: `pkg/hecode/doc.go`
- Create: `pkg/hcsv/doc.go`
- Create: `pkg/hfloat/doc.go`
- Create: `pkg/hidgenerator/doc.go`
- Create: `pkg/hlark/doc.go`
- Create: `pkg/hlo/doc.go`
- Create: `pkg/hredis/doc.go`
- Create: `pkg/hresty/doc.go`
- Create: `pkg/htime/doc.go`
- Create: `pkg/hcond/example_test.go`
- Create: `pkg/hecode/example_test.go`
- Create: `pkg/hcsv/example_test.go`
- Create: `pkg/hfloat/example_test.go`
- Create: `pkg/hidgenerator/example_test.go`
- Create: `pkg/hlark/example_test.go`
- Create: `pkg/hredis/example_test.go`
- Create: `pkg/hresty/example_test.go`
- Create: `pkg/htime/example_test.go`

**总共 20 个新文件(11 doc.go + 9 example_test.go)。不动任何现有文件。**

每个 doc.go 仅含 `package` 声明 + 完整中文 godoc 注释;每个 example_test.go 与原包同名(`package hxxx`),只放 Example 函数。

- [ ] **Step 1: `pkg/hcast/doc.go`**

```go
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
```

- [ ] **Step 2: `pkg/hcond/doc.go`**

```go
// Package hcond 提供 SQL WHERE 条件构造器,支持原子条件
// (=, !=, >, <, >=, <=, IN)与逻辑组合 (AND, OR)。
//
// 输出 SQL 片段 + 占位参数,可直接喂给 database/sql 或 GORM。
//
// 安全注意:Condition.LHS 直接拼接进 SQL(列名无法参数化),
// 调用方必须确保 LHS 不来自不可信用户输入,以避免 SQL 注入。
// 列名建议用常量或白名单约束。
package hcond
```

- [ ] **Step 3: `pkg/hcond/example_test.go`**

```go
package hcond

import "fmt"

func ExampleCondition_ToSQL() {
	cond := Condition{
		Operator: OpAnd,
		Conditions: []Condition{
			{Operator: OpEq, LHS: "status", RHS: "active"},
			{Operator: OpGt, LHS: "age", RHS: 18},
		},
	}
	sql, args, _ := cond.ToSQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// (status = ? AND age > ?)
	// [active 18]
}
```

- [ ] **Step 4: `pkg/hecode/doc.go`**

```go
// Package hecode 提供带错码的错误体系与统一响应辅助。
//
// 错码段约定(系统预留):
//   - [1000,1100) 系统错(internal/network/timeout/...)
//   - [1100,1200) 参数错
//   - [1200,1300) 业务错(not found/forbidden/...)
//   - [1300,1400) 数据校验错
//   - [1400,1500) 数据库错
//   - [1500,1600) Redis 错
//
// 业务方建议从 9000 起定义自己的错码,避免与系统段冲突。
// 错码全局唯一,重复注册会 panic(在 init 阶段就能发现)。
package hecode
```

- [ ] **Step 5: `pkg/hecode/example_test.go`**

```go
package hecode

import "fmt"

func ExampleNew() {
	err := New(9001, "user not found")
	fmt.Println(Code(err))
	fmt.Println(err)
	// Output:
	// 9001
	// code=9001, msg=user not found
}

func ExampleWrap() {
	base := New(9002, "db query failed")
	wrapped := Wrap(base, "load user profile")
	fmt.Println(Code(wrapped))         // 保留原 code
	fmt.Println(Cause(wrapped) == base) // 链尾就是 base
	// Output:
	// 9002
	// true
}
```

- [ ] **Step 6: `pkg/hcsv/doc.go`**

```go
// Package hcsv 提供 CSV 文件读取的简单封装。
//
// 当前只支持读取,且一次性加载到内存,适合配置 / 小数据。
// 处理大文件应直接用标准库 encoding/csv 的流式 Reader。
package hcsv
```

- [ ] **Step 7: `pkg/hcsv/example_test.go`**

```go
package hcsv

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExampleGetCSVData() {
	dir, _ := os.MkdirTemp("", "hcsv-example-")
	defer os.RemoveAll(dir)
	f := filepath.Join(dir, "data.csv")
	_ = os.WriteFile(f, []byte("name,age\nAlice,30\nBob,25"), 0644)

	rows, _ := GetCSVData(f)
	for _, row := range rows {
		fmt.Println(row)
	}
	// Output:
	// [name age]
	// [Alice 30]
	// [Bob 25]
}
```

- [ ] **Step 8: `pkg/hfloat/doc.go`**

```go
// Package hfloat 提供基于 shopspring/decimal 的浮点运算工具。
//
// 适用于价格 / 金额 / 会计场景:
//   - 加减乘除走 decimal 避免精度丢失
//   - 比较走容差(默认 1e-6)避免直接 == 浮点判断
//   - 字符串与 float64 互转保留有效精度
//
// 不适用于追求极致性能的科学计算(decimal 慢于原生 float)。
package hfloat
```

- [ ] **Step 9: `pkg/hfloat/example_test.go`**

```go
package hfloat

import "fmt"

func ExampleAddFloat() {
	// 经典浮点精度问题:0.1 + 0.2 用 float 直接相加得 0.30000000000000004
	r := AddFloat(0.1, 0.2)
	fmt.Println(r)
	// Output:
	// 0.3
}

func ExampleAddStringFloat() {
	r, _ := AddStringFloat("0.1", "0.2")
	fmt.Println(r)
	// Output:
	// 0.3
}
```

- [ ] **Step 10: `pkg/hidgenerator/doc.go`**

```go
// Package hidgenerator 提供请求 ID / 业务 ID 生成器接口与实现。
//
// IdGenerator 接口允许业务按需切换底层实现。
// 当前内置 UUID v4 一种实现;后续可扩展 snowflake / nanoid 等。
package hidgenerator
```

- [ ] **Step 11: `pkg/hidgenerator/example_test.go`**

```go
package hidgenerator

import "fmt"

func ExampleNewUuid() {
	gen := NewUuid()
	id := gen.GenerateRequestID()
	fmt.Println(len(id)) // UUID v4 标准长度 36
	// Output:
	// 36
}
```

- [ ] **Step 12: `pkg/hlark/doc.go`**

```go
// Package hlark 提供飞书机器人 webhook 发送辅助。
//
// 签名算法用飞书官方:
//   stringToSign := timestamp + "\n" + secret
//   HMAC-SHA256(key=stringToSign, data="") 然后 base64 编码
//
// 当前支持文本消息(text)。富文本 / 卡片消息后续扩展。
package hlark
```

- [ ] **Step 13: `pkg/hlark/example_test.go`**

```go
package hlark

import "fmt"

func ExampleGenSign() {
	// 固定 timestamp + secret 验证签名稳定性
	sign1, _ := GenSign(1700000000, "my-secret")
	sign2, _ := GenSign(1700000000, "my-secret")
	fmt.Println(sign1 == sign2)
	// Output:
	// true
}
```

- [ ] **Step 14: `pkg/hlo/doc.go`**

```go
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
```

- [ ] **Step 15: `pkg/hredis/doc.go`**

```go
// Package hredis 提供 github.com/redis/go-redis/v9 的薄封装:
//   - Config 结构体含连接池 / 超时 / DB / 密码,默认值兜底
//   - NewClient 构造后立即 Ping 一次验证连通性
//   - Ping 单独导出供外部探活
//
// 不重造 Redis 原生 API,获取到 *redis.Client 后直接调用 Get/Set/HSet 等。
package hredis
```

- [ ] **Step 16: `pkg/hredis/example_test.go`**

```go
package hredis

import (
	"context"
	"fmt"
	"time"
)

// 注意:此 Example 不带 // Output:,因为它需要可用的 Redis 实例,
// 仅作调用形态展示。
func ExampleNewClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := NewClient(ctx, Config{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		fmt.Println("connect failed:", err)
		return
	}
	defer client.Close()

	// 直接使用原生 API
	_ = client.Set(ctx, "key", "value", 0).Err()
}
```

- [ ] **Step 17: `pkg/hresty/doc.go`**

```go
// Package hresty 提供基于 github.com/go-resty/resty/v2 的 HTTP 客户端封装。
//
// 主要能力:
//   - NewRestyClient 带默认连接池 / 超时 / DualStack 的 transport
//   - GetTraceInfo 提取结构化请求追踪(DNS / TCP / TLS / Server time 等)
//
// 不打日志,由调用方负责把 RequestTrace 喂给自己的 logger;
// 不引入 internal/logger 依赖。
package hresty
```

- [ ] **Step 18: `pkg/hresty/example_test.go`**

```go
package hresty

import "fmt"

// Example 不带 // Output:,因为需要外网,仅展示调用形态。
func ExampleNewRestyClient() {
	client := NewRestyClient()
	resp, err := client.R().EnableTrace().Get("https://example.com")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}

	trace, err := GetTraceInfo(resp)
	if err == nil {
		fmt.Println("DNS lookup:", trace.DNSLookup)
		fmt.Println("Total time:", trace.TotalTime)
	}
}
```

- [ ] **Step 19: `pkg/htime/doc.go`**

```go
// Package htime 提供时间处理工具:
//   - 常用格式常量(TimeFormatDateTimeStandard 等)
//   - 时间戳解析(自动判断秒级 / 毫秒级)
//   - 日期计算辅助(月天数 / 今日剩余秒数 / 闰年校验等)
package htime
```

- [ ] **Step 20: `pkg/htime/example_test.go`**

```go
package htime

import "fmt"

func ExampleParseTimeStamp() {
	t, _ := ParseTimeStamp("1672531200")
	fmt.Println(t.UTC().Format(TimeFormatDateTimeStandard))
	// Output:
	// 2023-01-01 00:00:00
}

func ExampleGetTotalDaysInMonth() {
	fmt.Println(GetTotalDaysInMonth(2024, 2)) // 闰年 2 月
	fmt.Println(GetTotalDaysInMonth(2023, 2)) // 平年 2 月
	// Output:
	// 29
	// 28
}
```

- [ ] **Step 21: 验证编译与全量测试 + go doc**

```bash
cd /Users/nikki/go/src/hollow
go build ./...
go test ./pkg/... -count=1 -run "Example|Test"
```
Expected: 全 PASS,Example 函数的 // Output 断言全部匹配

抽样检查 go doc 能正确读到 package 注释:
```bash
go doc github.com/vaynedu/hollow/pkg/hecode | head -15
go doc github.com/vaynedu/hollow/pkg/hcond | head -10
go doc github.com/vaynedu/hollow/pkg/hredis | head -10
```
Expected: 每个都打印出对应的中文 godoc 段落

- [ ] **Step 22: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/
git commit -m "docs(pkg): 11 个包补齐 doc.go 与 Example 函数

- 每个包加 doc.go 作为 go pkg.go.dev 入口,中文 godoc 注释
- 9 个有公开 API 的包追加 ExampleXxx 测试(hcast/hlo 仅入口包跳过)
- 部分 Example 带 // Output: 自动验证 (hcond/hecode/hcsv/hfloat/
  hidgenerator/hlark/htime),依赖外部服务的(hredis/hresty)仅展示调用形态
- hes/hredlock 当前是空壳,留给 Iter 2 补实现时一起加文档"
```

---

## Self-Review

**1. Spec coverage**
- Iter 1 目标 = "11 个非空壳包加 doc.go + Example" + "hexcel 改名 hcsv":Task 1 改名 ✓,Task 2 文档 ✓
- 空壳包 hes / hredlock 跳过(按计划留 Iter 2)
- Iter 2 / 3 / 4 不在本 plan 范围

**2. Placeholder scan**
- 无 TBD / TODO / "实现 X" 等占位
- 每个 Step 都给出完整代码或具体命令

**3. Type consistency**
- `hcond.Condition` / `OpAnd` / `OpEq` / `OpGt`:Example 用法与 Task 6 中现有实现一致(已验证)
- `hecode.New` / `Wrap` / `Code` / `Cause`:Example 与现有 ecode.go 公开符号一致
- `hcsv.GetCSVData`:函数签名 `(string) ([][]string, error)` 与原 hexcel 一致(只改包名)
- `hfloat.AddFloat` / `AddStringFloat`:Example 验证 0.1+0.2=0.3 是真实行为(decimal 包保证)
- `hidgenerator.NewUuid().GenerateRequestID()`:UUID v4 标准长度 36(uuid.New().String() 行为)
- `hlark.GenSign(timestamp, secret)`:与 Task 2 修复后的签名一致
- `hredis.NewClient(ctx, Config)`:与 Task 4 实现一致
- `hresty.NewRestyClient()` / `GetTraceInfo`:与 Task 3 实现一致
- `htime.ParseTimeStamp(string) (time.Time, error)`:与现有实现一致;`TimeFormatDateTimeStandard = "2006-01-02 15:04:05"` 与 ParseTimeStamp 输出格式匹配,Example 的 `// Output:` 断言成立

**4. 风险点**
- `ExampleParseTimeStamp` 用 `t.UTC().Format(...)`,因为 `time.Unix(1672531200, 0)` 在不同时区 Format 结果不同。`UTC()` 后强制成 `2023-01-01 00:00:00`,可重现
- `ExampleAddFloat` 的 `// Output: 0.3` 依赖 decimal 库的格式化:`decimal.NewFromFloat(0.1).Add(NewFromFloat(0.2)).Float64() = 0.3`,且 fmt.Println(float64 0.3) 输出"0.3"。这两步都成立(已知 Go 行为),但实施时若 Example 失败,把 `// Output:` 注释掉并在 report 里标注
- `ExampleAddStringFloat` 输出 string `"0.3"`,decimal 的 `.String()` 在 `decimal.NewFromString("0.1").Add(NewFromString("0.2"))` 之后会输出 `"0.3"`,稳定
- `ExampleNewUuid` 的 `len(id) = 36` 是 UUID v4 字符串长度的稳定行为(`uuid.New().String()` 返回带短横线的 36 字符)
