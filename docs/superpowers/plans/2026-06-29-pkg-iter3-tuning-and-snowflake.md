# pkg Iter 3: 健壮性微调 + hidgenerator snowflake 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 Iter 1/2 累计的 Minor 与"健壮性微调"一次性扫尾,并给 hidgenerator 补 snowflake 实现(自实现,不引入新依赖)。

**Architecture:**
- **Task 1 文档微调**:README L91-95 排版修复 + 反映 hes/hredlock 已实现;hecode/doc.go 补 `Wrap` 非 EcodeError 时 ErrCodeUnknown=1099 兜底说明;hlark/doc.go 签名算法改 godoc 代码块格式
- **Task 2 代码微调**:hresty 删 `DualStack`(Go 1.12+ 已 deprecated);htime ParseTimeStamp 改用 strconv 判断秒/毫秒而非字符串长度;hfloat 补回精度损失边界用例;hes/example_test.go 修 nil panic 隐患
- **Task 3 hidgenerator snowflake**:64bit Twitter 标准实现(1+41+10+12),实现 IdGenerator 接口,同时暴露 GenerateInt64();不引入新依赖

每个 task 独立可提交,互不依赖。

**Tech Stack:** Go 1.25 / goconvey

## Global Constraints

- 包路径前缀 `github.com/vaynedu/hollow/pkg/...`
- pkg 不依赖 internal
- 测试统一 goconvey,不用 testify
- 中文 godoc + 中文 commit message
- Go 1.25
- 不引入任何新依赖
- 不动 cmd/, internal/, example/, examples/, hollow.go
- 不动 pkg/ 其它子目录(除明确列出的)
- 任务结束前 `go build ./... && go test ./pkg/... -count=1` 全绿
- interface{} → any 不强制改

## File Inventory

**Task 1(文档):**
- Modify: `README.md` — L91-95 排版 + hes/hredlock 描述
- Modify: `pkg/hecode/doc.go` — 补 Wrap 兜底说明
- Modify: `pkg/hlark/doc.go` — 签名算法改代码块格式

**Task 2(代码微调):**
- Modify: `pkg/hresty/resty.go` — 删 DualStack 字段
- Modify: `pkg/htime/time.go` — ParseTimeStamp 改 strconv 判断
- Modify: `pkg/htime/time_test.go` — 补充 strconv 路径测试
- Modify: `pkg/hfloat/float_test.go` — 补回精度损失边界用例
- Modify: `pkg/hes/example_test.go` — 修 nil panic 隐患

**Task 3(snowflake):**
- Create: `pkg/hidgenerator/snowflake.go`
- Create: `pkg/hidgenerator/snowflake_test.go`
- Modify: `pkg/hidgenerator/doc.go` — 加 snowflake 说明
- Modify: `pkg/hidgenerator/example_test.go` — 加 ExampleNewSnowflake

**不动:**
- pkg/ 其它子目录、cmd/、internal/、example/、examples/、hollow.go

---

## Task 1: 文档微调(README + hecode/hlark doc.go)

**目标:** 清理 3 个累计文档 Minor。

**Files:**
- Modify: `README.md`
- Modify: `pkg/hecode/doc.go`
- Modify: `pkg/hlark/doc.go`

- [ ] **Step 1: 改 `README.md` L91-95 反映 hes/hredlock 已实现**

当前(L91-95):
```
- 适用于秒杀等高并发场景 hcsv - CSV 文件读取
- 一次性加载 CSV 到内存,适合配置/小数据 hes - Elasticsearch 客户端
- ES 连接和操作封装
```

这段渲染时"上一项末尾 + 下一项标题"挤在一行(原 README 的 markdown 怪格式)。最小修复:让每个包标题独占一行,内容描述用次级 bullet。把现有的:
```
- 适用于秒杀等高并发场景 hcsv - CSV 文件读取
- 一次性加载 CSV 到内存,适合配置/小数据 hes - Elasticsearch 客户端
- ES 连接和操作封装
```

替换为:
```
- 适用于秒杀等高并发场景

hcsv - CSV 文件读取
- 一次性加载 CSV 到内存,适合配置/小数据

hes - Elasticsearch 客户端
- 基于 elasticsearch/v8,Config + NewClient(自带 Info 探活) + Ping 薄封装
- 拿到原生 *elasticsearch.Client 后直接使用 v8 esapi
```

同时把 L84-90 附近 `hredlock` 段(原描述"分布式锁 / 适用于秒杀等高并发场景")改为反映实际能力:
```
hredlock - 分布式锁
- 基于 go-redsync/v4,Lock/Unlock 二件套
- 接受外部 *redis.Client,锁 expiry 强制显式传入(防死锁)
```

**重要:** 这一段 README 整体格式很乱(Iter1-Fix 已经做过一次最小修复),本次只改 hredlock/hcsv/hes 三段的描述,**不要重排其它包描述**(htime/hlark 等)避免格式扩散。

如果 README 上下文不允许这样替换(比如 L84-90 不是 hredlock 段),按实际位置定位 hredlock/hcsv/hes 三段描述并就地替换为上面新版本。Implementer 在动手前 Read 完整 README 确认位置。

- [ ] **Step 2: 改 `pkg/hecode/doc.go` 末尾追加 Wrap 兜底说明**

原 doc.go(11 行)末尾追加一段(在 `package hecode` 之前):

```go
//
// Wrap / WithMessage 对非 EcodeError 类型的 err 会用 ErrCodeUnknown(1099)
// 作为兜底错码;EcodeError 类型的 err 则保留原 code。
```

最终 doc.go 内容:
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
//
// Wrap / WithMessage 对非 EcodeError 类型的 err 会用 ErrCodeUnknown(1099)
// 作为兜底错码;EcodeError 类型的 err 则保留原 code。
package hecode
```

- [ ] **Step 3: 改 `pkg/hlark/doc.go` 用代码块格式描述签名算法**

原 doc.go:
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

改为(关键:把签名算法两行的缩进从两空格改成 tab,godoc 才会渲染为等宽代码块):

```go
// Package hlark 提供飞书机器人 webhook 发送辅助。
//
// 签名算法用飞书官方:
//
//	stringToSign := timestamp + "\n" + secret
//	HMAC-SHA256(key=stringToSign, data="") 然后 base64 编码
//
// 当前支持文本消息(text)。富文本 / 卡片消息后续扩展。
package hlark
```

注意:godoc 把以 tab 开头的注释行渲染为代码块。原版用 3 个空格不会触发代码块格式。

- [ ] **Step 4: 验证**

```bash
cd /Users/nikki/go/src/hollow
go build ./...
go doc github.com/vaynedu/hollow/pkg/hecode | head -20
go doc github.com/vaynedu/hollow/pkg/hlark | head -15
```

Expected:
- build 全绿
- hecode 的 doc 输出末尾能看到 Wrap 兜底说明
- hlark 的 doc 输出中签名算法两行有等宽代码块缩进(以 tab 渲染)

- [ ] **Step 5: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add README.md pkg/hecode/doc.go pkg/hlark/doc.go
git commit -m "docs: 清理 README hes/hredlock 描述 + 补 hecode/hlark doc 细节

- README L91-95 反映 hes/hredlock 已实现的真实能力描述
- hecode/doc.go 补 Wrap 非 EcodeError 时 ErrCodeUnknown 兜底说明
- hlark/doc.go 签名算法用 tab 缩进, godoc 渲染为代码块"
```

---

## Task 2: 代码微调(hresty / htime / hfloat / hes Example)

**目标:** 清理 4 处累计的代码层 Minor。

**Files:**
- Modify: `pkg/hresty/resty.go`
- Modify: `pkg/htime/time.go`
- Modify: `pkg/htime/time_test.go`
- Modify: `pkg/hfloat/float_test.go`
- Modify: `pkg/hes/example_test.go`

- [ ] **Step 1: `pkg/hresty/resty.go` 删 DualStack**

当前 `NewTransport` 内 line 32 有:
```go
DialContext: (&net.Dialer{
    Timeout:   3 * time.Second,
    KeepAlive: 30 * time.Second,
    DualStack: true, // 是否启用 IPv4 和 IPv6 双栈支持(true 表示同时支持)
}).DialContext,
```

Go 1.12+ 后 `net.Dialer.DualStack` 已标 deprecated(`Deprecated: Functionality is now always enabled`),保留无害但应清掉。

**修法:** 删除 `DualStack: true,` 这一整行 + 同行尾注释。其它字段不动。

修改后:
```go
DialContext: (&net.Dialer{
    Timeout:   3 * time.Second,
    KeepAlive: 30 * time.Second,
}).DialContext,
```

- [ ] **Step 2: `pkg/htime/time.go` ParseTimeStamp 改 strconv 判断**

当前(L31-45):
```go
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
```

问题:用字符串长度判断秒/毫秒,边界脆(如 `"100"` 这种小时间戳长度不是 10 会被拒;`"1234567890123"` 长度 13 走毫秒分支但实际可能是秒)。

**修法:** 用 timeStamp 数值范围判断:
- 1e10 = 10^10 ≈ 2286-11-20 秒级,几乎不可能超过
- 1e10 以上视为毫秒级

新实现:
```go
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
```

注意:删掉了 `errors` 包的使用,如果文件 import 列表里 `errors` 已没人用,把 import 行一并删。当前 time.go 还有别处用 errors 吗?

让我看一眼:

```bash
grep -n "errors\." pkg/htime/time.go
```

如果只有 `ParseTimeStamp` 用,删 import;否则保留。

(implementer 自己确认。从我现有 context 看:`time.go` 完整内容 49 行,只有 ParseTimeStamp 用 `errors.New`,可以删 import。)

- [ ] **Step 3: `pkg/htime/time_test.go` 更新 ParseTimeStamp 测试**

新行为下,`"12345"`(5 位,5 < 1e10)应被视为秒级而非"invalid"。原测试用例 `"12345"` 期望错误,需要重新评估:

- 原"12345 视为 invalid"是因为长度判断;新逻辑下数值 12345 是秒级合法
- 实际语义:`"12345"` 解析成 time.Unix(12345, 0) = 1970-01-01 03:25:45 UTC
- 这是合法的(虽然时间很早),不应再报 invalid

**修法:** 删除 "InvalidTimestamp" 测试块,新增两个覆盖新行为的:

替换原 `TestParseTimeStamp` 函数为:

```go
func TestParseTimeStamp(t *testing.T) {
	Convey("ParseTimeStamp", t, func() {
		Convey("秒级 10 位", func() {
			got, err := ParseTimeStamp("1672531200")
			So(err, ShouldBeNil)
			So(got.Equal(time.Unix(1672531200, 0)), ShouldBeTrue)
		})
		Convey("毫秒级 13 位", func() {
			got, err := ParseTimeStamp("1672531200000")
			So(err, ShouldBeNil)
			So(got.Equal(time.UnixMilli(1672531200000)), ShouldBeTrue)
		})
		Convey("小数值视为秒级", func() {
			got, err := ParseTimeStamp("12345")
			So(err, ShouldBeNil)
			So(got.Equal(time.Unix(12345, 0)), ShouldBeTrue)
		})
		Convey("非数字字符串返回 error", func() {
			_, err := ParseTimeStamp("not-a-number")
			So(err, ShouldNotBeNil)
		})
	})
}
```

- [ ] **Step 4: `pkg/hfloat/float_test.go` 补回精度损失边界用例**

Iter 1 Task 6 缩水时丢了 4 个精度损失 case。在 `TestConvertFloatToString` 块内补回:

定位:`TestConvertFloatToString` 当前的 Convey 块,在 `cases` slice 内追加 4 个用例。完整新版本:

```go
func TestConvertFloatToString(t *testing.T) {
	Convey("ConvertFloatToString", t, func() {
		cases := []struct {
			in  float64
			out string
		}{
			{1.23, "1.23"},
			{1.0, "1"},
			{0.0, "0"},
			{0.000001, "0.000001"},
			{-1.23, "-1.23"},
			// 精度损失边界:float64 有效位约 15-17 位
			{123456789.123456789, "123456789.12345679"},
			{1234567890.123456789, "1234567890.1234567"},
			{12345678900.123456789, "12345678900.123457"},
			{123456789000.123456789, "123456789000.12346"},
		}
		for _, c := range cases {
			So(ConvertFloatToString(c.in), ShouldEqual, c.out)
		}
	})
}
```

注:精度边界 case 的期望值来自 Go strconv.FormatFloat 的实际行为(由 IEEE 754 双精度决定),Iter 1 Task 6 之前的原 test 已经验证过这些值。

- [ ] **Step 5: `pkg/hes/example_test.go` 修 nil panic 隐患**

当前 line 25-27:
```go
res, _ := client.Info(client.Info.WithContext(ctx))
defer res.Body.Close()
fmt.Println(res.Status())
```

问题:`client.Info(...)` 失败时 res 为 nil,`defer res.Body.Close()` 会 nil panic。Example 是给用户抄的模板,应该作示范。

修复后:
```go
res, err := client.Info(client.Info.WithContext(ctx))
if err != nil {
    fmt.Println("info failed:", err)
    return
}
defer res.Body.Close()
fmt.Println(res.Status())
```

- [ ] **Step 6: 验证编译 + 测试**

```bash
cd /Users/nikki/go/src/hollow
go build ./...
go test ./pkg/hresty/... ./pkg/htime/... ./pkg/hfloat/... ./pkg/hes/... -v -count=1
```

Expected:
- build 全绿
- htime 新增"小数值视为秒级"用例 PASS;原"12345 invalid"用例已被替换
- hfloat ConvertFloatToString 9 个 case 全 PASS(原 5 + 新 4)
- hresty / hes Example 不参与 // Output 断言,编译通过即可

- [ ] **Step 7: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hresty pkg/htime pkg/hfloat pkg/hes
git commit -m "refactor(pkg): 累计代码层微调清理

- hresty 删 net.Dialer.DualStack(Go 1.12+ 已 deprecated)
- htime ParseTimeStamp 改用数值范围判断秒/毫秒,不再依赖字符串长度
- hfloat float_test 补回 4 个 ConvertFloatToString 精度损失边界用例
- hes example_test 修 client.Info 失败 nil panic 隐患"
```

---

## Task 3: hidgenerator 加 Snowflake 实现

**目标:** hidgenerator 当前只有 UUID 一种实现,加 64bit Twitter 标准 snowflake(自实现,无新依赖)。

**Files:**
- Create: `pkg/hidgenerator/snowflake.go`
- Create: `pkg/hidgenerator/snowflake_test.go`
- Modify: `pkg/hidgenerator/doc.go` — 加 snowflake 说明
- Modify: `pkg/hidgenerator/example_test.go` — 加 ExampleNewSnowflake

**Interfaces produces:**
- `type Snowflake struct { ... }` — 实现 `IdGenerator` 接口
- `var ErrMachineIDOutOfRange = errors.New("hidgenerator: machineID must be in [0, 1024)")`
- `func NewSnowflake(machineID int64) (*Snowflake, error)`
- `func (s *Snowflake) GenerateRequestID() string` — 实现 IdGenerator,返回 base10 字符串
- `func (s *Snowflake) GenerateInt64() int64` — 直接拿 int64 ID

**位域布局(64bit):**
- bit 63: 符号位(总是 0)
- bit 62-22: 41 bit 毫秒级时间戳(相对自定义 epoch)
- bit 21-12: 10 bit 机器 ID(0-1023)
- bit 11-0: 12 bit 同毫秒内序列号(0-4095)

**自定义 epoch:** 2020-01-01 00:00:00 UTC = 1577836800000 ms;41 bit 容纳约 69 年,可用到 ~2089 年。

- [ ] **Step 1: 创建 `pkg/hidgenerator/snowflake.go`**

```go
package hidgenerator

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// Snowflake 64bit Twitter 标准实现:1 符号 + 41 毫秒时间戳 + 10 机器 ID + 12 序列
//
// 自定义 epoch:2020-01-01 UTC,可用到 ~2089 年
// 单机器单毫秒最多 4096 个 ID,理论峰值 ~4M/s
// 时钟回拨场景:简化处理,沿用上次时间戳防止 ID 倒退(可能导致序列号溢出 hang)
type Snowflake struct {
	mu        sync.Mutex
	epoch     int64 // 起始时间戳(毫秒)
	machineID int64 // 机器 ID,bit 21-12
	lastTime  int64 // 上次生成的毫秒时间戳
	sequence  int64 // 同毫秒内序列号
}

const (
	snowflakeEpoch       int64 = 1577836800000 // 2020-01-01 00:00:00 UTC ms
	snowflakeMachineBits uint  = 10
	snowflakeSeqBits     uint  = 12
	snowflakeMaxMachine  int64 = (1 << snowflakeMachineBits) - 1
	snowflakeMaxSeq      int64 = (1 << snowflakeSeqBits) - 1
)

// ErrMachineIDOutOfRange machineID 超出 [0, 1024) 范围时返回
var ErrMachineIDOutOfRange = errors.New("hidgenerator: machineID must be in [0, 1024)")

// NewSnowflake 构造 Snowflake 实例
// machineID 取值 [0, 1023],由部署方在集群内分配唯一值
func NewSnowflake(machineID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > snowflakeMaxMachine {
		return nil, ErrMachineIDOutOfRange
	}
	return &Snowflake{
		epoch:     snowflakeEpoch,
		machineID: machineID,
	}, nil
}

// GenerateRequestID 实现 IdGenerator 接口,返回 base10 字符串形式的 snowflake ID
func (s *Snowflake) GenerateRequestID() string {
	return strconv.FormatInt(s.GenerateInt64(), 10)
}

// GenerateInt64 直接生成 int64 形式的 snowflake ID,避免字符串分配
func (s *Snowflake) GenerateInt64() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.lastTime {
		// 时钟回拨,等到追回上次时间戳防止 ID 倒退
		now = s.lastTime
	}
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & snowflakeMaxSeq
		if s.sequence == 0 {
			// 序列溢出,等待下一毫秒
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}
	s.lastTime = now

	elapsed := now - s.epoch
	return (elapsed << (snowflakeMachineBits + snowflakeSeqBits)) |
		(s.machineID << snowflakeSeqBits) |
		s.sequence
}
```

- [ ] **Step 2: 创建 `pkg/hidgenerator/snowflake_test.go`**

```go
package hidgenerator

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSnowflake(t *testing.T) {
	Convey("NewSnowflake", t, func() {
		Convey("machineID 在范围内构造成功", func() {
			s, err := NewSnowflake(0)
			So(err, ShouldBeNil)
			So(s, ShouldNotBeNil)

			s2, err2 := NewSnowflake(1023)
			So(err2, ShouldBeNil)
			So(s2, ShouldNotBeNil)
		})

		Convey("machineID 负数返回 ErrMachineIDOutOfRange", func() {
			_, err := NewSnowflake(-1)
			So(err, ShouldEqual, ErrMachineIDOutOfRange)
		})

		Convey("machineID >= 1024 返回 ErrMachineIDOutOfRange", func() {
			_, err := NewSnowflake(1024)
			So(err, ShouldEqual, ErrMachineIDOutOfRange)
		})
	})
}

func TestSnowflakeGenerateInt64(t *testing.T) {
	Convey("Snowflake.GenerateInt64", t, func() {
		s, _ := NewSnowflake(1)

		Convey("ID 单调递增", func() {
			a := s.GenerateInt64()
			b := s.GenerateInt64()
			c := s.GenerateInt64()
			So(a, ShouldBeLessThan, b)
			So(b, ShouldBeLessThan, c)
		})

		Convey("ID 永远大于 0", func() {
			for i := 0; i < 100; i++ {
				So(s.GenerateInt64(), ShouldBeGreaterThan, 0)
			}
		})
	})
}

func TestSnowflakeGenerateRequestID(t *testing.T) {
	Convey("Snowflake.GenerateRequestID 实现 IdGenerator", t, func() {
		var gen IdGenerator
		gen, err := NewSnowflake(1)
		So(err, ShouldBeNil)

		id1 := gen.GenerateRequestID()
		id2 := gen.GenerateRequestID()
		So(id1, ShouldNotBeEmpty)
		So(id2, ShouldNotBeEmpty)
		So(id1, ShouldNotEqual, id2)
	})
}

func TestSnowflakeConcurrent(t *testing.T) {
	Convey("并发生成 ID 无重复", t, func() {
		s, _ := NewSnowflake(1)
		const goroutines = 10
		const perGoroutine = 1000

		var mu sync.Mutex
		ids := make(map[int64]struct{}, goroutines*perGoroutine)

		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				local := make([]int64, 0, perGoroutine)
				for j := 0; j < perGoroutine; j++ {
					local = append(local, s.GenerateInt64())
				}
				mu.Lock()
				for _, id := range local {
					ids[id] = struct{}{}
				}
				mu.Unlock()
			}()
		}
		wg.Wait()

		So(len(ids), ShouldEqual, goroutines*perGoroutine)
	})
}
```

注:并发测试用 `10 × 1000 = 10000` ID,远小于 4096/ms 上限,不会触发"等待下一毫秒"分支,测试用时应在 < 100ms 内。

- [ ] **Step 3: 修改 `pkg/hidgenerator/doc.go` 加 snowflake 说明**

原 doc.go(7 行)替换为:

```go
// Package hidgenerator 提供请求 ID / 业务 ID 生成器接口与实现。
//
// IdGenerator 接口允许业务按需切换底层实现。
// 当前内置两种实现:
//   - Uuid:基于 google/uuid v4,无状态,适合 request_id / trace_id
//   - Snowflake:64bit Twitter 标准实现,有状态(需 machineID),
//     单机峰值 ~4M/s,适合业务主键 ID
package hidgenerator
```

- [ ] **Step 4: 修改 `pkg/hidgenerator/example_test.go` 追加 ExampleNewSnowflake**

在文件末尾追加:

```go
func ExampleNewSnowflake() {
	gen, _ := NewSnowflake(1)
	id := gen.GenerateInt64()
	fmt.Println(id > 0)
	// Output:
	// true
}
```

注:不能断言具体 ID 值(依赖时间);只断言 > 0 验证基本正确性。

- [ ] **Step 5: 验证**

```bash
cd /Users/nikki/go/src/hollow
go build ./...
go test ./pkg/hidgenerator/... -v -count=1 -race
```

Expected:
- build 全绿
- 4 个 Test 全 PASS(NewSnowflake / GenerateInt64 / GenerateRequestID / Concurrent)
- ExampleNewSnowflake // Output 断言 PASS
- `-race` 不报数据竞争(因为有 mu 保护)

- [ ] **Step 6: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hidgenerator
git commit -m "feat(hidgenerator): 增加 Snowflake 实现

- 64bit Twitter 标准: 1 符号 + 41 毫秒时间戳 + 10 机器 ID + 12 序列
- epoch 2020-01-01 UTC, 单机峰值 ~4M/s, 可用到 ~2089 年
- 实现 IdGenerator 接口(GenerateRequestID 返 base10 字符串),
  同时暴露 GenerateInt64 避免字符串分配
- machineID 范围 [0, 1023], 越界返 ErrMachineIDOutOfRange
- 测试覆盖构造/单调递增/并发 10×1000 无重复, race 检测无冲突"
```

---

## Self-Review

**1. Spec coverage**
- 文档微调(README + hecode + hlark)→ Task 1 ✓
- 代码微调(hresty / htime / hfloat / hes Example)→ Task 2 ✓
- hidgenerator snowflake → Task 3 ✓
- hcond LHS 注释 — 已在 Iter 1 Task 2 加(verify 过),本 Iter 不重复
- Iter 4 新增工具包(hstr/hfile/hctx/hjson)不在范围

**2. Placeholder scan**
- 无 TBD / TODO 占位
- 每个 Step 给出完整代码或具体命令

**3. Type consistency**
- Snowflake 实现 IdGenerator 接口:`GenerateRequestID() string`,与 Uuid 类型一致;同时新增 `GenerateInt64() int64` 作为 Snowflake 专属方法
- htime ParseTimeStamp 签名 `(string) (time.Time, error)` 不变,只换实现
- 错误变量命名一致:`ErrMachineIDOutOfRange`、`ErrEmptyDSN` 等沿用 `Err...` 前缀

**4. 风险点**
- htime 改 strconv 判断后,小数值时间戳被视为秒级(原版会报 invalid),行为变化但更合理。Task 2 Step 3 同步更新测试覆盖新行为。
- hfloat 精度边界 case 期望值依赖 Go strconv.FormatFloat 行为,IEEE 754 双精度稳定。若某个 case 实际跑出来与预期略差(比如末位四舍五入),implementer 应以实际行为为准更新期望值并在 report 说明。
- Snowflake 测试 `10 × 1000` 远低于 4096/ms 上限,但理论上仍有时钟跳跃导致重复的风险(极低)。-race 主要验证 mu 保护。
- README 修改最复杂,implementer 在动手前应 Read 完整 README 确认上下文,避免误改其它段落。
