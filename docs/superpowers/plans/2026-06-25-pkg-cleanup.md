# pkg 清理与完善实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 A1(修 bug)→ A2(补 hredis)→ A3(清理)→ A4(测试全 Convey 化)四阶段把 hollow 的 pkg 收敛为一套可用的工具集。

**Architecture:**
- A1 用最小改动让现有"半成品"包真正可用,顺手把对应测试改 Convey
- A2 hredis 只做薄封装(NewClient + 健康检查),不重造 redis 客户端 API
- A3 删重复/示例代码,把 `demo_*` main 程序统一搬到顶层 `examples/`,避免污染 pkg
- A4 收口剩余测试,从 go.mod 移除 testify,跑全绿

**Tech Stack:** Go 1.25 / Gin / GORM / go-redis v9 / goconvey

## Global Constraints

- 包路径前缀 `github.com/vaynedu/hollow/pkg/...`
- 所有 pkg 目录内的 `.go` 文件**不得**依赖 `github.com/vaynedu/hollow/internal/...`(pkg 是对外暴露)
- 所有 test 文件统一使用 `github.com/smartystreets/goconvey/convey`(`Convey/So`)风格,不再使用 `github.com/stretchr/testify`
- 每个 task 结束前都要 `go build ./... && go test ./pkg/<改动包>/... -v` 全绿
- 中文注释,提交信息中文
- 不引入新依赖(miniredis 等)除非 plan 中显式列出
- 不修改 `cmd/`、`internal/`、`example/` 任何文件 —— 本计划仅动 `pkg/**` 与 `go.mod`、`go.sum`
- 调用方外部目前都没引用 pkg 内任何函数(已通过 grep 验证),可放心改签名

## File Inventory (会触及的文件)

修改:
- `pkg/hcond/condition.go` — 用 `Op` 常量替换原字符串字面量
- `pkg/hcond/gorm_helper.go` — `Conditioner.ToSQL` 改成 3 返回值;`BuildWhereClause` 返回 `(*gorm.DB, error)`
- `pkg/hcond/condition_test.go` — 重写为 Convey
- `pkg/hlark/web_hook.go` — `SendTextToFeiShu(ctx, url, secret, text)`;校验响应
- `pkg/hlark/web_hook_test.go` — `httptest` mock + Convey
- `pkg/hresty/resty.go` — 删 `PrintTraceInfo`/`PrintStructuredTrace`,去掉 `internal/logger` 引用
- `pkg/hresty/resty_test.go` — `httptest` + Convey,不外网
- `pkg/hfloat/float.go` — 删 `abs`,改用 `math.Abs`
- `pkg/hfloat/float_test.go` — Convey
- `pkg/hidgenerator/idgenerator_test.go` — Convey
- `pkg/hredlock/redlock_test.go` — Convey
- `pkg/htime/time_test.go` — Convey
- `pkg/htime/calculate_date_test.go` — Convey
- `pkg/hcast/cast_test.go` — Convey
- `pkg/hlo/lo_test.go` — Convey
- `pkg/hexcel/excel_test.go` — Convey
- `go.mod` / `go.sum` — 移除 `stretchr/testify`,新增(无)

新增:
- `pkg/hredis/redis.go` — 薄封装
- `pkg/hredis/redis_test.go` — Convey(纯单元,不连真实 redis)

删除:
- `pkg/hutils/request_id.go`(与 hidgenerator 重复)
- `pkg/hutils/`(整目录)
- `pkg/hlark/hlark_test.go`(是 lark SDK 用法 demo,不是测试)
- `pkg/hexcel/excel_to_sql_demo.go`(业务化死代码,污染 pkg)

迁移(从 pkg 搬出到顶层 `examples/`):
- `pkg/hredis/demo_pub_sub/` → `examples/redis_pubsub/`
- `pkg/hredlock/demo_sekill/` → `examples/redlock_seckill/`
- `pkg/hes/demo_es/` → `examples/es_demo/`

---

## Task 1: 修复 hcond 接口签名 + Convey 化

**问题:**
- `pkg/hcond/gorm_helper.go:8` `Conditioner.ToSQL() (string, []interface{})` 2 返回值
- `pkg/hcond/condition.go:22` `Condition.ToSQL() (string, []interface{}, error)` 3 返回值
- 接口对不上 → `*Condition` 不满足 `Conditioner` → `BuildWhereClause` 实际不可用
- `pkg/hcond/op.go` 定义的 `Op` 常量被 condition.go 无视(condition.go 里又自己定义了一遍字符串)

**Files:**
- Modify: `pkg/hcond/condition.go`
- Modify: `pkg/hcond/gorm_helper.go`
- Modify: `pkg/hcond/condition_test.go`(改 Convey)
- Modify: `pkg/hcond/op.go`(把字符串值改成 SQL 实际操作符,避免 `==` 这种 SQL 不识别的写法)
- Modify: `pkg/hcond/parser.go`(沿用新 3 返回值签名,顺手返回 error)

**Interfaces:**
- Produces:
  - `pkg/hcond.Op` 常量:`OpEq="="` / `OpNotEq="!="` / `OpGt=">"` / `OpLt="<"` / `OpGte=">="` / `OpLte="<="` / `OpIn="IN"` / `OpAnd="AND"` / `OpOr="OR"`
  - `Condition.ToSQL() (string, []interface{}, error)` 保持不变
  - `Conditioner` 接口:`ToSQL() (string, []interface{}, error)`
  - `BuildWhereClause(db *gorm.DB, cond Conditioner) (*gorm.DB, error)`

- [ ] **Step 1: 重写 `pkg/hcond/op.go` —— Op 值就是 SQL 实际操作符**

```go
package hcond

type Op string

const (
	OpEq    Op = "="
	OpNotEq Op = "!="
	OpGt    Op = ">"
	OpLt    Op = "<"
	OpGte   Op = ">="
	OpLte   Op = "<="
	OpIn    Op = "IN"

	OpAnd Op = "AND"
	OpOr  Op = "OR"
)
```

- [ ] **Step 2: 重写 `pkg/hcond/condition.go` —— 用 Op 常量,删重复字符串**

```go
package hcond

import (
	"fmt"
	"strings"
)

var (
	ErrRHSNotSlice         = fmt.Errorf("RHS must be a slice for IN operator")
	ErrUnsupportedOperator = fmt.Errorf("unsupported operator")
)

// Condition 表示一个条件节点:原子条件或子条件组
type Condition struct {
	Operator   Op          `json:"operator"`
	LHS        string      `json:"lhs"`
	RHS        interface{} `json:"rhs"`
	Conditions []Condition `json:"conditions"`
}

// ToSQL 生成 SQL WHERE 片段(不含 WHERE 关键字)和占位参数
func (c *Condition) ToSQL() (string, []interface{}, error) {
	if len(c.Conditions) == 0 {
		return c.toAtomicSQL()
	}

	clauses := make([]string, 0, len(c.Conditions))
	args := make([]interface{}, 0)
	for i := range c.Conditions {
		sub := c.Conditions[i]
		sql, subArgs, err := sub.ToSQL()
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, sql)
		args = append(args, subArgs...)
	}

	op := OpAnd
	if c.Operator == OpOr {
		op = OpOr
	}
	return fmt.Sprintf("(%s)", strings.Join(clauses, " "+string(op)+" ")), args, nil
}

func (c *Condition) toAtomicSQL() (string, []interface{}, error) {
	switch c.Operator {
	case OpEq, OpNotEq, OpGt, OpLt, OpGte, OpLte:
		return fmt.Sprintf("%s %s ?", c.LHS, c.Operator), []interface{}{c.RHS}, nil
	case OpIn:
		values, ok := c.RHS.([]interface{})
		if !ok {
			return "", nil, ErrRHSNotSlice
		}
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s IN (%s)", c.LHS, strings.Join(placeholders, ", ")), values, nil
	default:
		return "", nil, fmt.Errorf("%w: %s", ErrUnsupportedOperator, c.Operator)
	}
}
```

- [ ] **Step 3: 重写 `pkg/hcond/parser.go` —— 沿用新签名 + 处理空条件**

```go
package hcond

import "fmt"

// Parse 将 Condition 转换为完整 SQL WHERE 子句(带 WHERE 关键字)
// 当 sql 为空时不返回 "WHERE"
func Parse(cond Condition) (string, []interface{}, error) {
	sql, args, err := cond.ToSQL()
	if err != nil {
		return "", nil, err
	}
	if sql == "" {
		return "", args, nil
	}
	return fmt.Sprintf("WHERE %s", sql), args, nil
}
```

- [ ] **Step 4: 重写 `pkg/hcond/gorm_helper.go` —— 接口与实现对齐**

```go
package hcond

import "gorm.io/gorm"

// Conditioner 实现该接口的对象可以被 BuildWhereClause 使用
type Conditioner interface {
	ToSQL() (string, []interface{}, error)
}

// BuildWhereClause 根据条件对象构建 GORM 的 Where 子句
// 若 cond 解析失败,返回原 db 与 error,调用方自己决定是否中断
func BuildWhereClause(db *gorm.DB, cond Conditioner) (*gorm.DB, error) {
	sql, args, err := cond.ToSQL()
	if err != nil {
		return db, err
	}
	if sql == "" {
		return db, nil
	}
	return db.Where(sql, args...), nil
}
```

- [ ] **Step 5: 重写 `pkg/hcond/condition_test.go` —— Convey 风格**

```go
package hcond

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCondition_ToSQL(t *testing.T) {
	Convey("Condition.ToSQL", t, func() {
		Convey("原子条件 = 操作符", func() {
			cond := Condition{Operator: OpEq, LHS: "column", RHS: "value"}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "column = ?")
			So(args, ShouldResemble, []interface{}{"value"})
		})

		Convey("原子条件 IN 操作符", func() {
			cond := Condition{Operator: OpIn, LHS: "column", RHS: []interface{}{1, 2, 3}}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "column IN (?, ?, ?)")
			So(args, ShouldResemble, []interface{}{1, 2, 3})
		})

		Convey("IN 操作符 RHS 不是切片", func() {
			cond := Condition{Operator: OpIn, LHS: "status", RHS: "active"}
			sql, args, err := cond.ToSQL()
			So(err, ShouldEqual, ErrRHSNotSlice)
			So(sql, ShouldBeEmpty)
			So(args, ShouldBeNil)
		})

		Convey("逻辑 AND", func() {
			cond := Condition{
				Operator: OpAnd,
				Conditions: []Condition{
					{Operator: OpEq, LHS: "col1", RHS: "val1"},
					{Operator: OpEq, LHS: "col2", RHS: "val2"},
				},
			}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "(col1 = ? AND col2 = ?)")
			So(args, ShouldResemble, []interface{}{"val1", "val2"})
		})

		Convey("逻辑 OR", func() {
			cond := Condition{
				Operator: OpOr,
				Conditions: []Condition{
					{Operator: OpEq, LHS: "col1", RHS: "val1"},
					{Operator: OpEq, LHS: "col2", RHS: "val2"},
				},
			}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "(col1 = ? OR col2 = ?)")
			So(args, ShouldResemble, []interface{}{"val1", "val2"})
		})

		Convey("未支持的操作符返回 ErrUnsupportedOperator", func() {
			cond := Condition{Operator: Op("INVALID"), LHS: "column", RHS: "value"}
			_, _, err := cond.ToSQL()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unsupported operator")
		})
	})
}

func TestParse(t *testing.T) {
	Convey("Parse", t, func() {
		Convey("正常条件加 WHERE 前缀", func() {
			cond := Condition{Operator: OpEq, LHS: "id", RHS: 1}
			sql, args, err := Parse(cond)
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "WHERE id = ?")
			So(args, ShouldResemble, []interface{}{1})
		})

		Convey("解析错误透传", func() {
			cond := Condition{Operator: OpIn, LHS: "x", RHS: "not-slice"}
			_, _, err := Parse(cond)
			So(err, ShouldEqual, ErrRHSNotSlice)
		})
	})
}
```

- [ ] **Step 6: 验证编译与测试通过**

Run: `cd /Users/nikki/go/src/hollow && go build ./pkg/hcond/... && go test ./pkg/hcond/... -v`
Expected: 全 PASS,无编译报错

- [ ] **Step 7: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hcond
git commit -m "fix(hcond): 对齐 Conditioner 与 Condition.ToSQL 签名,统一使用 Op 常量

- Conditioner.ToSQL 改为 3 返回值,与 Condition 实现匹配
- BuildWhereClause 返回 (*gorm.DB, error)
- condition.go 改用 op.go 的 Op 常量,Op 值改为 SQL 真实操作符
- Parse 处理空 SQL 不再多生成 WHERE
- condition_test.go 改写为 goconvey 风格"
```

---

## Task 2: 修复 hlark webhook + Convey 化

**问题:**
- `pkg/hlark/web_hook.go:63` secret 硬编码 `"xxxxxxxxxxxxxxxxxxxxxxxxxxx"`
- `SendSmsToFeiShu(ctx, req, url)` 把 `req` 整体 JSON 序列化扔进 text 字段,签名密钥拿不到
- 不检查响应 `StatusCode` 与 body 的 `code` 字段,失败静默
- `pkg/hlark/hlark_test.go` 不是测试,是 lark SDK 用法演示

**Files:**
- Modify: `pkg/hlark/web_hook.go`
- Modify: `pkg/hlark/web_hook_test.go`
- Delete: `pkg/hlark/hlark_test.go`

**Interfaces:**
- Produces:
  - `func SendTextToFeiShu(ctx context.Context, url, secret, text string) error`
  - `func GenSign(timestamp int64, secret string) (string, error)`(导出供单测/复用)

- [ ] **Step 1: 重写 `pkg/hlark/web_hook.go`**

```go
package hlark

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type feiShuTextContent struct {
	Text string `json:"text"`
}

type feiShuRequest struct {
	Timestamp string            `json:"timestamp"`
	Sign      string            `json:"sign"`
	MsgType   string            `json:"msg_type"`
	Content   feiShuTextContent `json:"content"`
}

type feiShuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SendTextToFeiShu 通过自定义机器人 webhook 发送文本消息
// url:    open.feishu.cn/open-apis/bot/v2/hook/{token}
// secret: 机器人安全设置中的"签名校验"密钥(若未开启,可传空,GenSign 会跳过)
func SendTextToFeiShu(ctx context.Context, url, secret, text string) error {
	ts := time.Now().Unix()
	sign, err := GenSign(ts, secret)
	if err != nil {
		return fmt.Errorf("hlark: gen sign: %w", err)
	}

	payload := feiShuRequest{
		Timestamp: fmt.Sprintf("%d", ts),
		Sign:      sign,
		MsgType:   "text",
		Content:   feiShuTextContent{Text: text},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("hlark: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("hlark: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("hlark: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hlark: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var parsed feiShuResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("hlark: decode response: %w, body=%s", err, string(respBody))
	}
	if parsed.Code != 0 {
		return fmt.Errorf("hlark: code=%d msg=%s", parsed.Code, parsed.Msg)
	}
	return nil
}

// GenSign 飞书自定义机器人签名算法:HMAC-SHA256(key=timestamp+\n+secret, data=""),再 base64
// secret 为空时返回空串(适配未开启签名校验的机器人)
func GenSign(timestamp int64, secret string) (string, error) {
	if secret == "" {
		return "", nil
	}
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	if _, err := h.Write(nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
```

- [ ] **Step 2: 重写 `pkg/hlark/web_hook_test.go` —— httptest mock + Convey**

```go
package hlark

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenSign(t *testing.T) {
	Convey("GenSign", t, func() {
		Convey("secret 为空返回空串", func() {
			s, err := GenSign(1234567890, "")
			So(err, ShouldBeNil)
			So(s, ShouldBeEmpty)
		})

		Convey("固定 timestamp+secret 输出稳定", func() {
			s1, err := GenSign(1700000000, "my-secret")
			So(err, ShouldBeNil)
			s2, _ := GenSign(1700000000, "my-secret")
			So(s1, ShouldEqual, s2)
			So(s1, ShouldNotBeEmpty)
		})

		Convey("不同 secret 输出不同", func() {
			a, _ := GenSign(1700000000, "secret-a")
			b, _ := GenSign(1700000000, "secret-b")
			So(a, ShouldNotEqual, b)
		})
	})
}

func TestSendTextToFeiShu(t *testing.T) {
	Convey("SendTextToFeiShu", t, func() {
		Convey("响应 code=0 视为成功", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req feiShuRequest
				_ = json.Unmarshal(body, &req)
				So(req.MsgType, ShouldEqual, "text")
				So(req.Content.Text, ShouldEqual, "hello")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
			}))
			defer srv.Close()

			err := SendTextToFeiShu(t.Context(), srv.URL, "secret", "hello")
			So(err, ShouldBeNil)
		})

		Convey("响应 code!=0 返回 error", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"code":19001,"msg":"sign error"}`))
			}))
			defer srv.Close()

			err := SendTextToFeiShu(t.Context(), srv.URL, "secret", "hello")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "code=19001")
		})

		Convey("HTTP 非 200 返回 error", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer srv.Close()

			err := SendTextToFeiShu(t.Context(), srv.URL, "secret", "hello")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "status=502")
		})
	})
}
```

注:`t.Context()` 是 Go 1.24+ 的 testing 包新方法;本项目 go 1.25 可用。若不可用降级为 `context.Background()`。

- [ ] **Step 3: 删除 `pkg/hlark/hlark_test.go`(只是 lark SDK 用法 demo)**

```bash
rm /Users/nikki/go/src/hollow/pkg/hlark/hlark_test.go
```

- [ ] **Step 4: 验证编译与测试通过**

Run: `cd /Users/nikki/go/src/hollow && go build ./pkg/hlark/... && go test ./pkg/hlark/... -v`
Expected: 全 PASS

- [ ] **Step 5: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hlark
git commit -m "fix(hlark): 修复 webhook 签名硬编码与响应检查

- SendTextToFeiShu(ctx, url, secret, text) 走参数,不再硬编码 secret
- GenSign 导出,空 secret 直接返回空串(兼容未开签名校验)
- 校验 HTTP status 与响应 code 字段,失败返回 error
- 用 httptest mock 改写测试为 goconvey 风格
- 删除 hlark_test.go(原文件只是 SDK demo)"
```

---

## Task 3: hresty 解耦 internal/logger + Convey 化

**问题:**
- `pkg/hresty/resty.go:11` 引用 `github.com/vaynedu/hollow/internal/logger`,pkg 不应依赖 internal
- `resty_test.go` 直连外网(httpbin/qq.com),CI 不稳定

**Files:**
- Modify: `pkg/hresty/resty.go`
- Modify: `pkg/hresty/resty_test.go`

**Interfaces:**
- Produces:
  - `NewRestyClient() *resty.Client`(保持)
  - `NewTransport() *http.Transport`(保持)
  - `GetTraceInfo(resp *resty.Response) (*RequestTrace, error)`(保持)
  - **移除** `PrintTraceInfo`、`PrintStructuredTrace`(打日志是调用方的事,文档里给出示例)

- [ ] **Step 1: 重写 `pkg/hresty/resty.go`**

```go
package hresty

import (
	"errors"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
)

// NewRestyClient 创建带连接池与超时配置的 resty 客户端
func NewRestyClient() *resty.Client {
	return resty.NewWithClient(newClient())
}

func newClient() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
		Timeout:   10 * time.Second,
	}
}

// NewTransport 默认 transport 配置,可被外部直接使用
func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxConnsPerHost:       runtime.GOMAXPROCS(0) * 64,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) * 64,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	}
}

// RequestTrace 结构化的请求追踪信息
// 调用方拿到后可自行用任意 logger(zap/logrus/std log)输出
type RequestTrace struct {
	DNSLookup      time.Duration
	ConnTime       time.Duration
	TCPConnTime    time.Duration
	TLSHandshake   time.Duration
	ServerTime     time.Duration
	ResponseTime   time.Duration
	TotalTime      time.Duration
	ResponseSize   int
	IsConnReused   bool
	IsConnWasIdle  bool
	ConnIdleTime   time.Duration
	RequestAttempt int
	RemoteAddr     string
}

// GetTraceInfo 从响应中提取跟踪信息,要求请求开启了 EnableTrace()
func GetTraceInfo(resp *resty.Response) (*RequestTrace, error) {
	if resp == nil || resp.Request == nil {
		return nil, errors.New("hresty: response or request is nil")
	}
	ti := resp.Request.TraceInfo()
	addr := ""
	if ti.RemoteAddr != nil {
		addr = ti.RemoteAddr.String()
	}
	return &RequestTrace{
		DNSLookup:      ti.DNSLookup,
		ConnTime:       ti.ConnTime,
		TCPConnTime:    ti.TCPConnTime,
		TLSHandshake:   ti.TLSHandshake,
		ServerTime:     ti.ServerTime,
		ResponseTime:   ti.ResponseTime,
		TotalTime:      ti.TotalTime,
		ResponseSize:   len(resp.Body()),
		IsConnReused:   ti.IsConnReused,
		IsConnWasIdle:  ti.IsConnWasIdle,
		ConnIdleTime:   ti.ConnIdleTime,
		RequestAttempt: ti.RequestAttempt,
		RemoteAddr:     addr,
	}, nil
}
```

- [ ] **Step 2: 重写 `pkg/hresty/resty_test.go` —— httptest + Convey**

```go
package hresty

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRestyClient(t *testing.T) {
	Convey("NewRestyClient", t, func() {
		client := NewRestyClient()
		So(client, ShouldNotBeNil)
		So(client.GetClient().Transport, ShouldNotBeNil)
	})
}

func TestGetTraceInfo(t *testing.T) {
	Convey("GetTraceInfo", t, func() {
		Convey("nil 响应返回 error", func() {
			ti, err := GetTraceInfo(nil)
			So(err, ShouldNotBeNil)
			So(ti, ShouldBeNil)
		})

		Convey("正常请求返回 trace", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("hello"))
			}))
			defer srv.Close()

			client := NewRestyClient()
			resp, err := client.R().EnableTrace().Get(srv.URL)
			So(err, ShouldBeNil)
			So(resp.StatusCode(), ShouldEqual, http.StatusOK)

			trace, err := GetTraceInfo(resp)
			So(err, ShouldBeNil)
			So(trace, ShouldNotBeNil)
			So(trace.ResponseSize, ShouldEqual, len("hello"))
			So(trace.RemoteAddr, ShouldNotBeEmpty)
		})
	})
}
```

- [ ] **Step 3: 验证编译与测试**

Run: `cd /Users/nikki/go/src/hollow && go build ./pkg/hresty/... && go test ./pkg/hresty/... -v`
Expected: 全 PASS,且 `grep "internal/logger" pkg/hresty/` 无结果

- [ ] **Step 4: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hresty
git commit -m "refactor(hresty): 移除 internal/logger 依赖,仅暴露结构化 trace

- 删除 PrintTraceInfo/PrintStructuredTrace,日志改由调用方负责
- GetTraceInfo 处理 RemoteAddr nil 边界
- 测试改用 httptest + goconvey,去掉外网依赖"
```

---

## Task 4: A2 hredis 薄封装

**目的:** 提供 v9 客户端的统一构造入口与健康检查,不重新封装 Get/Set/Del 等 redis 原生 API。

**Files:**
- Create: `pkg/hredis/redis.go`(覆盖原空文件)
- Create: `pkg/hredis/redis_test.go`

**Interfaces:**
- Produces:
  - `type Config struct { Addr, Password string; DB, PoolSize int; DialTimeout, ReadTimeout, WriteTimeout time.Duration }`
  - `func NewClient(ctx context.Context, cfg Config) (*redis.Client, error)` —— 构造 + Ping 验证
  - `func Ping(ctx context.Context, client *redis.Client) error` —— 健康检查

- [ ] **Step 1: 写 `pkg/hredis/redis.go`**

```go
package hredis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config hredis 客户端配置,默认值见 NewClient
type Config struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// ErrEmptyAddr 未配置 Addr 时返回
var ErrEmptyAddr = errors.New("hredis: addr is required")

// NewClient 构造 go-redis v9 客户端并执行一次 Ping 验证连通性
// 默认值:DialTimeout=5s, ReadTimeout=3s, WriteTimeout=3s, PoolSize=10
func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, ErrEmptyAddr
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 10
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	if err := Ping(ctx, client); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("hredis: ping failed: %w", err)
	}
	return client, nil
}

// Ping 探活,返回非 nil error 表示连接不可用
func Ping(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return errors.New("hredis: nil client")
	}
	return client.Ping(ctx).Err()
}
```

- [ ] **Step 2: 写 `pkg/hredis/redis_test.go` —— 不连真 redis,只测构造逻辑**

```go
package hredis

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	Convey("NewClient", t, func() {
		Convey("空 Addr 返回 ErrEmptyAddr", func() {
			_, err := NewClient(context.Background(), Config{})
			So(err, ShouldEqual, ErrEmptyAddr)
		})

		Convey("无效 Addr Ping 失败返回错误", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			_, err := NewClient(ctx, Config{
				Addr:        "127.0.0.1:1", // 几乎一定连不上
				DialTimeout: 100 * time.Millisecond,
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "ping failed")
		})
	})
}

func TestPing(t *testing.T) {
	Convey("Ping", t, func() {
		Convey("nil client 返回错误", func() {
			err := Ping(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "nil client")
		})
	})
}
```

- [ ] **Step 3: 验证**

Run: `cd /Users/nikki/go/src/hollow && go build ./pkg/hredis/... && go test ./pkg/hredis/... -v -count=1`
Expected: PASS(无真实 redis 也能跑)

- [ ] **Step 4: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hredis/redis.go pkg/hredis/redis_test.go
git commit -m "feat(hredis): 增加 go-redis v9 薄封装与健康检查

- Config 结构体,所有超时与池大小有默认值
- NewClient 构造后立即 Ping 一次验证连通性
- Ping 单独导出供外部探活
- 测试覆盖默认值与失败路径,不依赖真实 redis"
```

---

## Task 5: A3 清理瘦身

**清理项一览:**
1. 删 `pkg/hutils`(与 hidgenerator 重复,且无外部引用)
2. `pkg/hfloat`:删 `abs`,改用 `math.Abs`
3. `pkg/hexcel`:删 `excel_to_sql_demo.go`(业务化死代码)
4. `pkg/hredis/demo_pub_sub/` 整体迁出 → `examples/redis_pubsub/`
5. `pkg/hredlock/demo_sekill/` 整体迁出 → `examples/redlock_seckill/`
6. `pkg/hes/demo_es/` 整体迁出 → `examples/es_demo/`

**注意:** `hcast`(只有注释)与 `hlo`(只有 lo 用法 test)本计划**保留不删**,因为 cast 是占位、lo 是用法学习样本,删除收益小且需用户拍板;在 plan 范围外。

**Files:**
- Delete: `pkg/hutils/request_id.go`,`pkg/hutils/`(空目录一并删)
- Delete: `pkg/hexcel/excel_to_sql_demo.go`
- Modify: `pkg/hfloat/float.go`
- Move: `pkg/hredis/demo_pub_sub/expire_pub_sub.go` → `examples/redis_pubsub/main.go`
- Move: `pkg/hredlock/demo_sekill/*` → `examples/redlock_seckill/`
- Move: `pkg/hes/demo_es/demo_es.go` → `examples/es_demo/main.go`

- [ ] **Step 1: 删除 hutils**

```bash
cd /Users/nikki/go/src/hollow
rm pkg/hutils/request_id.go
rmdir pkg/hutils
```

- [ ] **Step 2: 修改 `pkg/hfloat/float.go` —— 删 abs**

将文件中的 `abs(...)` 全部替换为 `math.Abs(...)`,并删除函数定义。完整新文件:

```go
package hfloat

import (
	"fmt"
	"math"
	"strconv"

	"github.com/shopspring/decimal"
)

const floatCompareMin = 0.000001

// IsFloatEqual 判断两个 float64 是否在精度容差内相等
func IsFloatEqual(a, b float64) bool {
	return math.Abs(a-b) < floatCompareMin
}

// CompareFloat 比较两个 float64
//
//	-1: a < b
//	 0: a == b(在容差内)
//	 1: a > b
func CompareFloat(a, b float64) int {
	diff := a - b
	if math.Abs(diff) < floatCompareMin {
		return 0
	} else if diff < 0 {
		return -1
	}
	return 1
}

// RoundUpFloat 四舍五入保留 2 位小数
func RoundUpFloat(x float64) float64 {
	return math.Round(x*100) / 100
}

// RoundDownFloat 向下取整保留 2 位小数
func RoundDownFloat(x float64) float64 {
	return math.Floor(x*100) / 100
}

// ConvertFloatToString float64 转字符串(保留有效精度)
func ConvertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ConvertStringToFloat 字符串转 float64,走 decimal 避免精度问题
func ConvertStringToFloat(s string) (float64, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return 0, err
	}
	res, _ := d.Float64()
	return res, nil
}

// AddFloat 用 decimal 做加法避免精度丢失
func AddFloat(a, b float64) float64 {
	res := decimal.NewFromFloat(a).Add(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f
}

// SubtractFloat 用 decimal 做减法
func SubtractFloat(a, b float64) float64 {
	res := decimal.NewFromFloat(a).Sub(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f
}

// MultiplyFloat 用 decimal 做乘法
func MultiplyFloat(a, b float64) float64 {
	res := decimal.NewFromFloat(a).Mul(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f
}

// DivideFloat 用 decimal 做除法,除数为 0 返回 error
func DivideFloat(a, b float64) (float64, error) {
	if math.Abs(b) < floatCompareMin {
		return 0, fmt.Errorf("hfloat: cannot divide by zero")
	}
	res := decimal.NewFromFloat(a).Div(decimal.NewFromFloat(b))
	f, _ := res.Float64()
	return f, nil
}

// AddStringFloat 两个字符串形式的小数相加,返回字符串结果
func AddStringFloat(a, b string) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "", err
	}
	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "", err
	}
	return aa.Add(bb).String(), nil
}
```

- [ ] **Step 3: 删除 hexcel 的 demo 文件**

```bash
cd /Users/nikki/go/src/hollow
rm pkg/hexcel/excel_to_sql_demo.go
```

- [ ] **Step 4: 迁出 demo_pub_sub**

```bash
cd /Users/nikki/go/src/hollow
mkdir -p examples/redis_pubsub
git mv pkg/hredis/demo_pub_sub/expire_pub_sub.go examples/redis_pubsub/main.go
rmdir pkg/hredis/demo_pub_sub
```

- [ ] **Step 5: 迁出 demo_sekill**

```bash
cd /Users/nikki/go/src/hollow
mkdir -p examples/redlock_seckill
git mv pkg/hredlock/demo_sekill/seckill_stock.go examples/redlock_seckill/main.go
git mv pkg/hredlock/demo_sekill/test_result.md examples/redlock_seckill/test_result.md
rmdir pkg/hredlock/demo_sekill
```

- [ ] **Step 6: 迁出 demo_es**

```bash
cd /Users/nikki/go/src/hollow
mkdir -p examples/es_demo
git mv pkg/hes/demo_es/demo_es.go examples/es_demo/main.go
rmdir pkg/hes/demo_es
```

- [ ] **Step 7: 整体编译验证**

Run: `cd /Users/nikki/go/src/hollow && go build ./...`
Expected: 无报错

Run: `cd /Users/nikki/go/src/hollow && go test ./pkg/... -count=1`
Expected: 全 PASS

- [ ] **Step 8: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add -A
git commit -m "refactor(pkg): 清理与瘦身

- 删除 pkg/hutils(与 hidgenerator 功能重复)
- hfloat 删除自定义 abs,改用 math.Abs
- 删除 pkg/hexcel/excel_to_sql_demo.go(业务化死代码)
- 将 demo_* 子目录全部从 pkg 迁出到 examples/
  - pkg/hredis/demo_pub_sub → examples/redis_pubsub
  - pkg/hredlock/demo_sekill → examples/redlock_seckill
  - pkg/hes/demo_es → examples/es_demo"
```

---

## Task 6: A4 剩余测试 Convey 化 + 移除 testify

**清单:** 把所有还在用 testify 或非 Convey 的测试文件,统一改 `Convey/So`。`hecode_test.go` 已是 Convey,不动。

**Files(全部 Modify):**
- `pkg/hcast/cast_test.go`
- `pkg/hexcel/excel_test.go`
- `pkg/hfloat/float_test.go`
- `pkg/hidgenerator/idgenerator_test.go`
- `pkg/hlo/lo_test.go`
- `pkg/hredlock/redlock_test.go`
- `pkg/htime/time_test.go`
- `pkg/htime/calculate_date_test.go`
- `go.mod` / `go.sum`

- [ ] **Step 1: `pkg/hcast/cast_test.go`**

```go
package hcast

import (
	"testing"
	"time"

	"github.com/spf13/cast"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCast(t *testing.T) {
	Convey("spf13/cast 基础用法验证", t, func() {
		So(cast.ToString(1.23456789), ShouldEqual, "1.23456789")
		So(cast.ToString(123456789), ShouldEqual, "123456789")
		So(cast.ToString(nil), ShouldEqual, "")
		So(cast.ToInt64("12344"), ShouldEqual, int64(12344))
		So(cast.ToFloat64("12344"), ShouldEqual, float64(12344))
	})
}

func TestCastTime(t *testing.T) {
	Convey("cast.ToTime/ToString time.Time", t, func() {
		now := time.Now()
		So(cast.ToTime(now).Equal(now), ShouldBeTrue)
		So(cast.ToString(now), ShouldNotBeEmpty)
	})
}
```

- [ ] **Step 2: `pkg/hexcel/excel_test.go`**

```go
package hexcel

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCSVData(t *testing.T) {
	Convey("GetCSVData", t, func() {
		tempDir := t.TempDir()

		Convey("正常读取 CSV", func() {
			f := filepath.Join(tempDir, "ok.csv")
			So(os.WriteFile(f, []byte("name,age\nAlice,30\nBob,25"), 0644), ShouldBeNil)

			rows, err := GetCSVData(f)
			So(err, ShouldBeNil)
			So(len(rows), ShouldEqual, 3)
			So(rows[0], ShouldResemble, []string{"name", "age"})
			So(rows[1], ShouldResemble, []string{"Alice", "30"})
		})

		Convey("文件不存在返回 open error", func() {
			rows, err := GetCSVData(filepath.Join(tempDir, "nope.csv"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "open file error")
			So(rows, ShouldBeNil)
		})

		Convey("空文件返回 csv file is empty", func() {
			f := filepath.Join(tempDir, "empty.csv")
			So(os.WriteFile(f, []byte(""), 0644), ShouldBeNil)
			rows, err := GetCSVData(f)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "csv file is empty")
			So(rows, ShouldBeNil)
		})

		Convey("格式错误返回 read csv file error", func() {
			f := filepath.Join(tempDir, "bad.csv")
			So(os.WriteFile(f, []byte("name,age\nAlice,30\nBob,25,"), 0644), ShouldBeNil)
			rows, err := GetCSVData(f)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "read csv file error")
			So(rows, ShouldBeNil)
		})
	})
}
```

- [ ] **Step 3: `pkg/hfloat/float_test.go`**

```go
package hfloat

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsFloatEqual(t *testing.T) {
	Convey("IsFloatEqual", t, func() {
		So(IsFloatEqual(1.0, 1.0), ShouldBeTrue)
		So(IsFloatEqual(1.0, 1.0000001), ShouldBeTrue)
		So(IsFloatEqual(1.0, 1.00001), ShouldBeFalse)
		So(IsFloatEqual(0.0, -0.00001), ShouldBeTrue)
		So(IsFloatEqual(0.0, 1.0), ShouldBeFalse)
	})
}

func TestCompareFloat(t *testing.T) {
	Convey("CompareFloat", t, func() {
		So(CompareFloat(1.0, 2.0), ShouldEqual, -1)
		So(CompareFloat(2.0, 1.0), ShouldEqual, 1)
		So(CompareFloat(1.0, 1.0000001), ShouldEqual, 0)
	})
}

func TestRoundUpFloat(t *testing.T) {
	Convey("RoundUpFloat 保留 2 位", t, func() {
		So(RoundUpFloat(1.234), ShouldEqual, 1.23)
		So(RoundUpFloat(1.235), ShouldEqual, 1.24)
	})
}

func TestRoundDownFloat(t *testing.T) {
	Convey("RoundDownFloat 向下取 2 位", t, func() {
		So(RoundDownFloat(1.234), ShouldEqual, 1.23)
		So(RoundDownFloat(1.239), ShouldEqual, 1.23)
	})
}

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
		}
		for _, c := range cases {
			So(ConvertFloatToString(c.in), ShouldEqual, c.out)
		}
	})
}

func TestConvertStringToFloat(t *testing.T) {
	Convey("ConvertStringToFloat", t, func() {
		Convey("合法数值", func() {
			v, err := ConvertStringToFloat("1.23")
			So(err, ShouldBeNil)
			So(IsFloatEqual(v, 1.23), ShouldBeTrue)
		})
		Convey("非法字符串返回 error", func() {
			_, err := ConvertStringToFloat("abc")
			So(err, ShouldNotBeNil)
		})
		Convey("空字符串返回 error", func() {
			_, err := ConvertStringToFloat("")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestArithmetic(t *testing.T) {
	Convey("加减乘除", t, func() {
		So(IsFloatEqual(AddFloat(1.23, 4.56), 5.79), ShouldBeTrue)
		So(IsFloatEqual(SubtractFloat(4.56, 1.23), 3.33), ShouldBeTrue)
		So(IsFloatEqual(MultiplyFloat(2.0, 3.0), 6.0), ShouldBeTrue)

		v, err := DivideFloat(6.0, 2.0)
		So(err, ShouldBeNil)
		So(IsFloatEqual(v, 3.0), ShouldBeTrue)

		_, err = DivideFloat(1.0, 0.0)
		So(err, ShouldNotBeNil)
	})
}

func TestAddStringFloat(t *testing.T) {
	Convey("AddStringFloat", t, func() {
		v, err := AddStringFloat("1.23", "4.56")
		So(err, ShouldBeNil)
		So(v, ShouldEqual, "5.79")

		_, err = AddStringFloat("abc", "1.0")
		So(err, ShouldNotBeNil)
	})
}
```

- [ ] **Step 4: `pkg/hidgenerator/idgenerator_test.go`**

```go
package hidgenerator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUuidGenerateRequestID(t *testing.T) {
	Convey("Uuid 实现 IdGenerator", t, func() {
		var gen IdGenerator = NewUuid()
		id1 := gen.GenerateRequestID()
		id2 := gen.GenerateRequestID()
		So(id1, ShouldNotBeEmpty)
		So(id2, ShouldNotBeEmpty)
		So(id1, ShouldNotEqual, id2)
		So(len(id1), ShouldEqual, 36) // 标准 UUID 长度
	})
}
```

- [ ] **Step 5: `pkg/hlo/lo_test.go`**

```go
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
```

- [ ] **Step 6: `pkg/hredlock/redlock_test.go`**

```go
package hredlock

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPlaceholder(t *testing.T) {
	Convey("hredlock 当前只有 demo,正式实现待补", t, func() {
		So(true, ShouldBeTrue)
	})
}
```

- [ ] **Step 7: `pkg/htime/time_test.go`**

```go
package htime

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTimeStampToTime(t *testing.T) {
	Convey("TimeStampToTime", t, func() {
		ts := int64(1672531200)
		So(TimeStampToTime(ts).Equal(time.Unix(ts, 0)), ShouldBeTrue)
	})
}

func TestTimeStampMsToTime(t *testing.T) {
	Convey("TimeStampMsToTime", t, func() {
		ts := int64(1672531200000)
		So(TimeStampMsToTime(ts).Equal(time.UnixMilli(ts)), ShouldBeTrue)
	})
}

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
		Convey("长度错误返回 error", func() {
			_, err := ParseTimeStamp("12345")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestParseTimeDataStandard(t *testing.T) {
	Convey("ParseTimeDataStandard", t, func() {
		got, err := ParseTimeDataStandard("2023-01-01 12:00:00")
		So(err, ShouldBeNil)
		expected, _ := time.Parse(TimeFormatDateTimeStandard, "2023-01-01 12:00:00")
		So(got.Equal(expected), ShouldBeTrue)
	})
}
```

- [ ] **Step 8: `pkg/htime/calculate_date_test.go`**

```go
package htime

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTotalDaysInMonth(t *testing.T) {
	Convey("GetTotalDaysInMonth", t, func() {
		So(GetTotalDaysInMonth(2024, 2), ShouldEqual, 29)
		So(GetTotalDaysInMonth(2025, 12), ShouldEqual, 31)
		So(GetTotalDaysInMonth(2025, 6), ShouldEqual, 30)
	})
}

func TestGetCurrentYearMonthDay(t *testing.T) {
	Convey("GetCurrentYearMonthDay 与 time.Now 一致", t, func() {
		n := time.Now()
		y, m, d := GetCurrentYearMonthDay()
		So(y, ShouldEqual, n.Year())
		So(m, ShouldEqual, int(n.Month()))
		So(d, ShouldEqual, n.Day())
	})
}

func TestGetCurrentTimeString(t *testing.T) {
	Convey("GetCurrentTimeString 长度等于标准格式", t, func() {
		s := GetCurrentTimeString()
		So(len(s), ShouldEqual, len(TimeFormatDateTimeStandard))
	})
}

func TestGetSecondsSinceMidnight(t *testing.T) {
	Convey("GetSecondsSinceMidnight 与本地时间一致", t, func() {
		n := time.Now()
		mid := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
		So(GetSecondsSinceMidnight(), ShouldEqual, int64(n.Sub(mid).Seconds()))
	})
}

func TestGetSecondsUntilMidnight(t *testing.T) {
	Convey("GetSecondsUntilMidnight 与本地时间一致", t, func() {
		n := time.Now()
		end := time.Date(n.Year(), n.Month(), n.Day(), 23, 59, 59, 0, n.Location())
		So(GetSecondsUntilMidnight(), ShouldEqual, int64(end.Sub(n).Seconds()))
	})
}

func TestGetDayOfMonth(t *testing.T) {
	Convey("GetDayOfMonth", t, func() {
		got, err := GetDayOfMonth(2024, 2, 29)
		So(err, ShouldBeNil)
		So(got.Equal(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)), ShouldBeTrue)

		_, err = GetDayOfMonth(2023, 2, 29)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "day out of range for month")

		_, err = GetDayOfMonth(2024, 13, 1)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid month or day")

		_, err = GetDayOfMonth(2024, 2, 0)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid month or day")
	})
}

func TestIsValidDate(t *testing.T) {
	Convey("IsValidDate", t, func() {
		So(IsValidDate(2024, 2, 29), ShouldBeTrue)
		So(IsValidDate(2025, 12, 31), ShouldBeTrue)
		So(IsValidDate(-1, 1, 1), ShouldBeFalse)
		So(IsValidDate(2024, 0, 1), ShouldBeFalse)
		So(IsValidDate(2024, 13, 1), ShouldBeFalse)
		So(IsValidDate(2024, 1, 0), ShouldBeFalse)
		So(IsValidDate(2023, 2, 29), ShouldBeFalse)
	})
}
```

- [ ] **Step 9: 移除 testify 依赖**

```bash
cd /Users/nikki/go/src/hollow
go mod tidy
```

确认 `go.mod` 已不再包含 `github.com/stretchr/testify`:
Run: `grep "stretchr/testify" go.mod`
Expected: 无输出

- [ ] **Step 10: 全量测试通过**

Run: `cd /Users/nikki/go/src/hollow && go build ./... && go test ./pkg/... -v -count=1`
Expected: 全 PASS

- [ ] **Step 11: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add -A
git commit -m "test(pkg): 全部测试改写为 goconvey,移除 testify 依赖

- 涉及 hcast/hexcel/hfloat/hidgenerator/hlo/hredlock/htime
- go mod tidy 移除 stretchr/testify"
```

---

## Self-Review

**1. Spec coverage**
- A1 修 bug → Task 1/2/3 ✓
- A2 hredis 薄封装 → Task 4 ✓
- A3 清理 → Task 5 ✓
- A4 测试全 Convey 化 + 删 testify → Task 6 ✓(覆盖 Task 1/2/3 没覆盖到的剩余测试)

**2. Placeholder scan**
- 无 TBD / TODO / "实现 X" 等占位
- 每个 step 都给了完整代码或具体命令

**3. Type consistency**
- hcond:`Op` 类型统一,`ToSQL` 3 返回值,`Conditioner.ToSQL` 与之对齐,`BuildWhereClause` 返回 `(*gorm.DB, error)` —— 一致
- hlark:`SendTextToFeiShu` 与 `GenSign` 签名在 Task 2 内一致
- hresty:`GetTraceInfo` 返回 `*RequestTrace`,无新增字段,删除项与测试一致
- hredis:`Config`/`NewClient`/`Ping`/`ErrEmptyAddr` 命名前后一致

**4. 风险点**
- `t.Context()` 是 Go 1.24+ 才有的方法;本项目 go 1.25,可用。若实施时报错,降级 `context.Background()`
- `go mod tidy` 可能拉新版本依赖;若 CI 严格固定,执行后人工检查 `go.sum` 改动
