# hgorm 薄封装 + 错误响应分级映射 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 hollow 框架补两个生产必需能力 —— GORM 单实例薄封装(连接池 + 慢 SQL 日志),以及 hecode 错码到 HTTP status 的分级映射(业务错 200 / 系统错 5xx)。

**Architecture:**
- `pkg/hgorm` 新增包,与 `pkg/hredis` 形态对齐(Config + NewDB + Ping,不重造 GORM API)。慢 SQL 日志通过实现 GORM 的 `logger.Interface` 接到 zap。
- `pkg/hecode` 新增 `ToHTTPStatus(err)` 函数,按错码段分类(1xxx 系统 / 14xx db / 15xx redis → 500;11xx 参数 / 12xx 业务 / 13xx 数据 → 200;非 EcodeError → 500)。
- `internal/middleware/reponse.go` 重命名为 `response.go`(修拼写),改用 `hecode.ToHTTPStatus` + `hecode.Code` 决定 HTTP code 与 body code。

**Tech Stack:** Go 1.25 / GORM v1.30 / go.uber.org/zap / goconvey

## Global Constraints

- 包路径前缀 `github.com/vaynedu/hollow/...`
- pkg 内不得依赖 `internal/...`(hgorm 自带 logger,不引 internal/logger)
- 测试统一 goconvey,不引入 testify
- 中文注释 + 中文 commit message
- Go 1.25
- 不本任务收紧 `interface{}` → `any` 风格
- 不引入新的第三方依赖(GORM/zap 已在 go.mod)
- 不动 cmd/, example/, examples/ 任何文件
- 任务结束前 `go build ./... && go test ./... -count=1` 全绿

## File Inventory

**新增:**
- `pkg/hgorm/gorm.go` — Config + NewDB + Ping + ErrEmptyDSN
- `pkg/hgorm/logger.go` — 自定义 GORM logger.Interface 实现,慢 SQL 走 zap.Warn
- `pkg/hgorm/gorm_test.go` — Convey 单元测试

**修改:**
- `pkg/hecode/ecode.go` — 新增 `ToHTTPStatus(err error) int`
- `pkg/hecode/ecode_test.go` — 追加 ToHTTPStatus 的 Convey 测试块
- `internal/middleware/reponse.go` → 重命名为 `internal/middleware/response.go` 并改用 hecode 映射
- `internal/middleware/middleware_test.go` — 更新 ResponseMiddleware 相关测试(若有)以覆盖新行为
- `internal/middleware/builtin.go` — 无改动(`NewResponseMiddleware()` 函数名不变)

**不动:**
- `internal/middleware/{request_id,logging,recovery,metrics,middleware,builtin}.go`
- `pkg/` 其它子目录
- `hollow.go`、`cmd/`、`example/`、`examples/`

---

## Task 1: pkg/hgorm 单实例薄封装

**Files:**
- Create: `pkg/hgorm/gorm.go`
- Create: `pkg/hgorm/logger.go`
- Create: `pkg/hgorm/gorm_test.go`

**Interfaces produces:**
- `type Config struct { DSN, Dialect string; MaxOpenConns, MaxIdleConns int; ConnMaxLifetime, ConnMaxIdleTime, SlowThreshold time.Duration; LogLevel string }`
- `var ErrEmptyDSN = errors.New("hgorm: dsn is required")`
- `var ErrUnsupportedDialect = errors.New("hgorm: unsupported dialect")`
- `func NewDB(ctx context.Context, cfg Config, log *zap.Logger) (*gorm.DB, error)`
- `func Ping(ctx context.Context, db *gorm.DB) error`
- `func NewZapGormLogger(log *zap.Logger, slow time.Duration, level gormlogger.LogLevel) gormlogger.Interface`

**Dialect 范围:** 本任务只支持 `mysql` 与 `postgres`,且**不引入新驱动依赖**;`Dialect` 字段当前作为校验字段,实际打开使用 `gorm.io/driver/mysql` 与 `gorm.io/driver/postgres` —— 这两个驱动**不在 go.mod**,因此 NewDB 实现里只引 `gorm.io/gorm` 并接收一个 `gorm.Dialector` 类型的可选 override。具体见 Step 2。

**简化(避免新增依赖):**
- 不直接 import 任何驱动包
- 新增 `func NewDBWithDialector(ctx, dialector gorm.Dialector, cfg Config, log *zap.Logger) (*gorm.DB, error)`:调用方自己构造 dialector(`mysql.Open(dsn)` / `postgres.Open(dsn)`),hgorm 只做共用部分(logger、连接池、Ping)
- `NewDB(ctx, cfg, log)` 失败返回 `ErrUnsupportedDialect`(占位,要求 v2 引入驱动后再实现)—— 或者本任务直接**不提供 NewDB**,只提供 `NewDBWithDialector`,文档示例展示如何调

**最终决定:** 只提供 `NewDBWithDialector`。理由:不引新依赖、给业务方完全自主选择驱动版本的空间。`ErrEmptyDSN` 改为在 dialector 为 nil 时返回 `ErrNilDialector`。

修正后 interfaces:
- `var ErrNilDialector = errors.New("hgorm: dialector is required")`
- `func NewDB(ctx context.Context, dialector gorm.Dialector, cfg Config, log *zap.Logger) (*gorm.DB, error)`
- `func Ping(ctx context.Context, db *gorm.DB) error`
- `func NewZapGormLogger(log *zap.Logger, slow time.Duration, level gormlogger.LogLevel) gormlogger.Interface`

- [ ] **Step 1: 写 `pkg/hgorm/gorm.go`**

```go
// Package hgorm 提供 GORM 单实例薄封装:连接池配置、慢 SQL 日志、健康检查。
// 不重造 GORM API,获取到 *gorm.DB 后调用方可直接使用 GORM 原生方法。
//
// 设计上不 import 任何具体数据库驱动(mysql/postgres/sqlite),调用方自行构造
// gorm.Dialector 传入,以便业务方自主选择驱动与版本。示例:
//
//	import "gorm.io/driver/mysql"
//	dialector := mysql.Open("user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true")
//	db, err := hgorm.NewDB(ctx, dialector, hgorm.Config{...}, logger)
package hgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config hgorm 连接配置,默认值见 NewDB 注释
type Config struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
	// LogLevel: "silent" | "error" | "warn" | "info",默认 "warn"
	LogLevel string `mapstructure:"log_level"`
}

// ErrNilDialector dialector 为 nil 时返回
var ErrNilDialector = errors.New("hgorm: dialector is required")

// NewDB 构造 *gorm.DB 并完成连接池配置 + 慢 SQL 日志接入 + Ping 验证连通性
//
// 默认值:
//   - MaxOpenConns=50, MaxIdleConns=10
//   - ConnMaxLifetime=1h, ConnMaxIdleTime=30m
//   - SlowThreshold=200ms
//   - LogLevel="warn"
func NewDB(ctx context.Context, dialector gorm.Dialector, cfg Config, log *zap.Logger) (*gorm.DB, error) {
	if dialector == nil {
		return nil, ErrNilDialector
	}
	if log == nil {
		log = zap.NewNop()
	}
	applyDefaults(&cfg)

	gormLogger := NewZapGormLogger(log, cfg.SlowThreshold, parseLogLevel(cfg.LogLevel))
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("hgorm: open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("hgorm: get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := Ping(ctx, db); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("hgorm: ping failed: %w", err)
	}
	return db, nil
}

// Ping 通过底层 sql.DB 探活
func Ping(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("hgorm: nil db")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("hgorm: get sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

func applyDefaults(cfg *Config) {
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime <= 0 {
		cfg.ConnMaxLifetime = time.Hour
	}
	if cfg.ConnMaxIdleTime <= 0 {
		cfg.ConnMaxIdleTime = 30 * time.Minute
	}
	if cfg.SlowThreshold <= 0 {
		cfg.SlowThreshold = 200 * time.Millisecond
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "warn"
	}
}

func parseLogLevel(s string) gormlogger.LogLevel {
	switch s {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}
```

- [ ] **Step 2: 写 `pkg/hgorm/logger.go` —— GORM logger 接 zap**

```go
package hgorm

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// zapGormLogger 实现 gorm.io/gorm/logger.Interface,把 GORM 日志桥接到 zap
type zapGormLogger struct {
	log           *zap.Logger
	slowThreshold time.Duration
	level         gormlogger.LogLevel
}

// NewZapGormLogger 构造接 zap 的 GORM logger
// slow 表示慢 SQL 阈值;level 控制最低输出级别
func NewZapGormLogger(log *zap.Logger, slow time.Duration, level gormlogger.LogLevel) gormlogger.Interface {
	if log == nil {
		log = zap.NewNop()
	}
	if slow <= 0 {
		slow = 200 * time.Millisecond
	}
	return &zapGormLogger{log: log, slowThreshold: slow, level: level}
}

func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *zapGormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		l.log.Sugar().Infof(msg, data...)
	}
}

func (l *zapGormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.log.Sugar().Warnf(msg, data...)
	}
}

func (l *zapGormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		l.log.Sugar().Errorf(msg, data...)
	}
}

// Trace 记录单条 SQL 执行,自动判断:
//   - 出错且非 RecordNotFound 且 level >= Error → Error
//   - 超过慢阈值且 level >= Warn → Warn(慢 SQL)
//   - level >= Info → Info(正常 SQL)
func (l *zapGormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.log.Error("gorm trace error", append(fields, zap.Error(err))...)
	case elapsed > l.slowThreshold && l.slowThreshold > 0 && l.level >= gormlogger.Warn:
		l.log.Warn("gorm slow sql",
			append(fields, zap.Duration("threshold", l.slowThreshold))...)
	case l.level >= gormlogger.Info:
		l.log.Info("gorm trace", fields...)
	}
}
```

- [ ] **Step 3: 写 `pkg/hgorm/gorm_test.go` —— Convey,不依赖真实数据库**

```go
package hgorm

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestApplyDefaults(t *testing.T) {
	Convey("applyDefaults 填充零值", t, func() {
		cfg := Config{}
		applyDefaults(&cfg)
		So(cfg.MaxOpenConns, ShouldEqual, 50)
		So(cfg.MaxIdleConns, ShouldEqual, 10)
		So(cfg.ConnMaxLifetime, ShouldEqual, time.Hour)
		So(cfg.ConnMaxIdleTime, ShouldEqual, 30*time.Minute)
		So(cfg.SlowThreshold, ShouldEqual, 200*time.Millisecond)
		So(cfg.LogLevel, ShouldEqual, "warn")
	})

	Convey("applyDefaults 不覆盖已设置值", t, func() {
		cfg := Config{
			MaxOpenConns:    100,
			MaxIdleConns:    20,
			ConnMaxLifetime: 2 * time.Hour,
			ConnMaxIdleTime: time.Hour,
			SlowThreshold:   500 * time.Millisecond,
			LogLevel:        "info",
		}
		applyDefaults(&cfg)
		So(cfg.MaxOpenConns, ShouldEqual, 100)
		So(cfg.MaxIdleConns, ShouldEqual, 20)
		So(cfg.ConnMaxLifetime, ShouldEqual, 2*time.Hour)
		So(cfg.ConnMaxIdleTime, ShouldEqual, time.Hour)
		So(cfg.SlowThreshold, ShouldEqual, 500*time.Millisecond)
		So(cfg.LogLevel, ShouldEqual, "info")
	})
}

func TestParseLogLevel(t *testing.T) {
	Convey("parseLogLevel", t, func() {
		So(parseLogLevel("silent"), ShouldEqual, gormlogger.Silent)
		So(parseLogLevel("error"), ShouldEqual, gormlogger.Error)
		So(parseLogLevel("warn"), ShouldEqual, gormlogger.Warn)
		So(parseLogLevel("info"), ShouldEqual, gormlogger.Info)
		So(parseLogLevel(""), ShouldEqual, gormlogger.Warn) // 默认
		So(parseLogLevel("unknown"), ShouldEqual, gormlogger.Warn)
	})
}

func TestNewDB(t *testing.T) {
	Convey("NewDB", t, func() {
		Convey("nil dialector 返回 ErrNilDialector", func() {
			_, err := NewDB(context.Background(), nil, Config{}, zap.NewNop())
			So(err, ShouldEqual, ErrNilDialector)
		})
	})
}

func TestPing(t *testing.T) {
	Convey("Ping", t, func() {
		Convey("nil db 返回错误", func() {
			err := Ping(context.Background(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "nil db")
		})
	})
}

func TestZapGormLogger(t *testing.T) {
	Convey("zapGormLogger", t, func() {
		log := zap.NewNop()

		Convey("NewZapGormLogger nil logger 自动用 Nop", func() {
			l := NewZapGormLogger(nil, 100*time.Millisecond, gormlogger.Info)
			So(l, ShouldNotBeNil)
		})

		Convey("NewZapGormLogger slow<=0 用默认 200ms", func() {
			l := NewZapGormLogger(log, 0, gormlogger.Info).(*zapGormLogger)
			So(l.slowThreshold, ShouldEqual, 200*time.Millisecond)
		})

		Convey("LogMode 返回新实例,原对象不变", func() {
			l := NewZapGormLogger(log, 100*time.Millisecond, gormlogger.Warn).(*zapGormLogger)
			l2 := l.LogMode(gormlogger.Error).(*zapGormLogger)
			So(l.level, ShouldEqual, gormlogger.Warn)
			So(l2.level, ShouldEqual, gormlogger.Error)
		})

		Convey("Trace silent 级别不 panic", func() {
			l := NewZapGormLogger(log, 100*time.Millisecond, gormlogger.Silent)
			So(func() {
				l.Trace(context.Background(), time.Now(),
					func() (string, int64) { return "SELECT 1", 1 }, nil)
			}, ShouldNotPanic)
		})

		Convey("Trace RecordNotFound 不当 error 上报", func() {
			l := NewZapGormLogger(log, 100*time.Millisecond, gormlogger.Error)
			So(func() {
				l.Trace(context.Background(), time.Now(),
					func() (string, int64) { return "SELECT 1", 0 }, gorm.ErrRecordNotFound)
			}, ShouldNotPanic)
		})
	})
}
```

- [ ] **Step 4: 验证编译与测试**

Run: `cd /Users/nikki/go/src/hollow && go build ./... && go test ./pkg/hgorm/... -v -count=1`
Expected: 全 PASS,无新增依赖(`go mod tidy` 后 go.mod/go.sum 不应有改动)

Run: `cd /Users/nikki/go/src/hollow && go mod tidy && git diff --stat go.mod go.sum`
Expected: 无 diff(GORM 已在 go.mod,本任务不引入新依赖)

- [ ] **Step 5: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hgorm
git commit -m "feat(hgorm): 增加 GORM 单实例薄封装

- Config 含连接池/慢 SQL 阈值/日志级别,默认值兜底
- NewDB 接收用户构造的 gorm.Dialector,不引入驱动依赖
- ZapGormLogger 实现 GORM logger.Interface,慢 SQL 走 zap.Warn
- Ping 单独导出供外部探活
- 测试覆盖 Config 默认值/parseLogLevel/nil 边界/logger trace 路径"
```

---

## Task 2: hecode 错码 → HTTP status 映射 + ResponseMiddleware 改造

**Files:**
- Modify: `pkg/hecode/ecode.go` — 新增 `ToHTTPStatus`
- Modify: `pkg/hecode/ecode_test.go` — 追加 Convey 测试块
- Delete: `internal/middleware/reponse.go`(拼写错误)
- Create: `internal/middleware/response.go`(正名 + 改造)
- Modify: `internal/middleware/middleware_test.go` — 更新 ResponseMiddleware 相关测试覆盖新行为

**Interfaces produces:**
- `pkg/hecode.ToHTTPStatus(err error) int`
  - err == nil → 200
  - 非 EcodeError → 500
  - 系统错段 1001~1099 → 500
  - 参数错段 1100~1199 → 200(用户语义"业务错 200")
  - 业务错段 1200~1299 → 200
  - 数据校验段 1300~1399 → 200
  - 数据库段 1400~1499 → 500
  - Redis 段 1500~1599 → 500
  - 其它已注册段(未来扩展)默认 → 200(业务错)

**映射规则(常量段):**
```
[1000, 1100) 系统错        → 500
[1100, 1400) 参数/业务/数据 → 200
[1400, 1600) db/redis     → 500
其它非空段                  → 200
```

**响应行为(改造前 → 改造后):**

| 场景 | 改造前 | 改造后 |
|---|---|---|
| handler 正常 `c.Set("data", x)` | 200 + code=200 + msg=success | 200 + code=0 + msg=success + data=x(保持 ok) |
| handler `c.Error(EcodeError 业务)` | 500 + code=500 + msg=err.Error() | 200 + code=hecode.Code + msg=hecode.GetMessage |
| handler `c.Error(EcodeError 系统/db/redis)` | 500 + code=500 + msg=err.Error() | 5xx(对应映射) + code=hecode.Code + msg=hecode.GetMessage |
| handler `c.Error(普通 error)` | 500 + code=500 + msg=err.Error() | 500 + code=1099(ErrCodeUnknown) + msg=err.Error() |

注:成功响应的 `Code` 字段由 `200` 改成 `0`(与 pkg/hecode/response.go 的 `Success` 函数对齐,避免一份代码两份"成功语义")。

- [ ] **Step 1: 在 `pkg/hecode/ecode.go` 末尾新增 `ToHTTPStatus`**

```go
// ToHTTPStatus 按错码段返回 HTTP 状态码
//
// 段位规则:
//   - nil                  → 200
//   - 非 EcodeError        → 500
//   - [1000,1100) 系统错   → 500
//   - [1100,1400) 业务错   → 200(参数/业务/数据校验,前端用 body code 判断)
//   - [1400,1600) db/redis → 500
//   - 其它已知错码          → 200(默认按业务错处理)
func ToHTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	code := Code(err)
	if code == 0 {
		// 非 EcodeError 或解包后无 code
		return 500
	}
	switch {
	case code >= 1000 && code < 1100:
		return 500
	case code >= 1100 && code < 1400:
		return 200
	case code >= 1400 && code < 1600:
		return 500
	default:
		return 200
	}
}
```

- [ ] **Step 2: 追加 `pkg/hecode/ecode_test.go` 的 ToHTTPStatus 测试块**

在文件末尾追加(`TestEdgeCases` 之后):

```go
// 测试 ToHTTPStatus 错码段映射
func TestToHTTPStatus(t *testing.T) {
	Convey("ToHTTPStatus", t, func() {
		Convey("nil 返 200", func() {
			So(ToHTTPStatus(nil), ShouldEqual, 200)
		})

		Convey("非 EcodeError 返 500", func() {
			So(ToHTTPStatus(errors.New("plain")), ShouldEqual, 500)
		})

		Convey("系统错段 1001~1099 → 500", func() {
			So(ToHTTPStatus(ErrInternal), ShouldEqual, 500)   // 1001
			So(ToHTTPStatus(ErrUnknown), ShouldEqual, 500)    // 1099
		})

		Convey("参数/业务/数据段 1100~1399 → 200", func() {
			So(ToHTTPStatus(ErrInvalidParam), ShouldEqual, 200)   // 1100
			So(ToHTTPStatus(ErrNotFound), ShouldEqual, 200)       // 1200
			So(ToHTTPStatus(ErrDataValidation), ShouldEqual, 200) // 1300
		})

		Convey("数据库段 1400~1499 → 500", func() {
			So(ToHTTPStatus(ErrDatabase), ShouldEqual, 500)        // 1400
			So(ToHTTPStatus(ErrDBTransactionCommit), ShouldEqual, 500) // 1417
		})

		Convey("Redis 段 1500~1599 → 500", func() {
			So(ToHTTPStatus(ErrRedisConnection), ShouldEqual, 500) // 1500
		})

		Convey("其它已知段 → 200", func() {
			err := New(9000, "custom") // [1600, +∞) 默认 200
			So(ToHTTPStatus(err), ShouldEqual, 200)
		})
	})
}
```

- [ ] **Step 3: 删除 `internal/middleware/reponse.go`**

```bash
cd /Users/nikki/go/src/hollow
git rm internal/middleware/reponse.go
```

- [ ] **Step 4: 创建 `internal/middleware/response.go` —— 用 hecode 映射**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hecode"
)

// ResponseMiddleware 自动封装响应:
//   - 正常路径:取 c.Get("data") 包装为 {code:0, msg:"success", data}
//   - handler 调过 c.Error(err):按 hecode.ToHTTPStatus 决定 HTTP code,
//     body 用 hecode 的 code/msg
type ResponseMiddleware struct{}

func NewResponseMiddleware() *ResponseMiddleware {
	return &ResponseMiddleware{}
}

func (m *ResponseMiddleware) HandlerFunc() gin.HandlerFunc {
	return responseMiddleware
}

func (m *ResponseMiddleware) Identifier() string {
	return "response"
}

// Response 标准响应格式
type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func responseMiddleware(c *gin.Context) {
	c.Next()

	// RequestID 由 RequestIDMiddleware 已写入 header,这里只读
	requestID := c.Writer.Header().Get("X-Request-ID")

	// handler 主动 c.Error(err) 的错误优先返回
	if e := c.Errors.Last(); e != nil {
		err := e.Err
		httpStatus := hecode.ToHTTPStatus(err)
		c.JSON(httpStatus, Response{
			Code:      errCode(err),
			Msg:       errMsg(err),
			RequestID: requestID,
		})
		return
	}

	// 正常路径:从 context 取 data
	data, _ := c.Get("data")
	c.JSON(200, Response{
		Code:      0,
		Msg:       "success",
		RequestID: requestID,
		Data:      data,
	})
}

// errCode 优先取 hecode 的 code,非 EcodeError 返 ErrCodeUnknown(1099)
func errCode(err error) int {
	if c := hecode.Code(err); c != 0 {
		return c
	}
	return hecode.ErrCodeUnknown
}

// errMsg 优先取 EcodeError 的 msg,fallback 到 err.Error()
func errMsg(err error) string {
	type messager interface{ GetMessage() string }
	if m, ok := err.(messager); ok {
		return m.GetMessage()
	}
	return err.Error()
}
```

注: `errMsg` 使用接口断言而非具体类型,避免循环引用且支持其它实现 `GetMessage()` 的类型。

- [ ] **Step 5: 更新 `internal/middleware/middleware_test.go` —— 给 ResponseMiddleware 加新行为测试**

先读现有文件,在末尾追加(原有 4 个测试保留,只追加 ResponseMiddleware 的覆盖测试):

```go
func TestResponseMiddleware_Success(t *testing.T) {
	Convey("ResponseMiddleware 成功路径返 code=0 + data", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(NewResponseMiddleware().HandlerFunc())
		router.GET("/ok", func(c *gin.Context) {
			c.Set("data", gin.H{"foo": "bar"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ok", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 200)
		So(w.Body.String(), ShouldContainSubstring, `"code":0`)
		So(w.Body.String(), ShouldContainSubstring, `"msg":"success"`)
		So(w.Body.String(), ShouldContainSubstring, `"foo":"bar"`)
	})
}

func TestResponseMiddleware_BusinessError(t *testing.T) {
	Convey("ResponseMiddleware 业务错(11xx)返 HTTP 200 + body code", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(NewResponseMiddleware().HandlerFunc())
		router.GET("/biz-err", func(c *gin.Context) {
			_ = c.Error(hecode.ErrInvalidParam)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/biz-err", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 200)
		So(w.Body.String(), ShouldContainSubstring, `"code":1100`)
		So(w.Body.String(), ShouldContainSubstring, "invalid parameter")
	})
}

func TestResponseMiddleware_SystemError(t *testing.T) {
	Convey("ResponseMiddleware 系统错(1xxx)返 HTTP 500 + body code", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(NewResponseMiddleware().HandlerFunc())
		router.GET("/sys-err", func(c *gin.Context) {
			_ = c.Error(hecode.ErrInternal)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/sys-err", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 500)
		So(w.Body.String(), ShouldContainSubstring, `"code":1001`)
	})
}

func TestResponseMiddleware_DBError(t *testing.T) {
	Convey("ResponseMiddleware 数据库错(14xx)返 HTTP 500", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(NewResponseMiddleware().HandlerFunc())
		router.GET("/db-err", func(c *gin.Context) {
			_ = c.Error(hecode.ErrDatabase)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/db-err", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 500)
		So(w.Body.String(), ShouldContainSubstring, `"code":1400`)
	})
}

func TestResponseMiddleware_PlainError(t *testing.T) {
	Convey("ResponseMiddleware 非 EcodeError 返 500 + code=1099", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(NewResponseMiddleware().HandlerFunc())
		router.GET("/plain-err", func(c *gin.Context) {
			_ = c.Error(errors.New("plain"))
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/plain-err", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 500)
		So(w.Body.String(), ShouldContainSubstring, `"code":1099`)
		So(w.Body.String(), ShouldContainSubstring, "plain")
	})
}
```

import 块需要加 `"errors"` 与 `"github.com/vaynedu/hollow/pkg/hecode"`。

- [ ] **Step 6: 验证编译与测试**

Run: `cd /Users/nikki/go/src/hollow && go build ./... && go test ./... -count=1`
Expected: 全 PASS,无新增依赖

- [ ] **Step 7: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hecode internal/middleware
git commit -m "feat(hecode): 增加 ToHTTPStatus 错码段映射,ResponseMiddleware 改用之

- hecode.ToHTTPStatus: 1xxx 系统/14xx db/15xx redis → 500,其它 → 200
- 文件名修正 reponse.go → response.go
- ResponseMiddleware:
  * 成功路径 code 由 200 改为 0,与 pkg/hecode/response.go 的 Success 对齐
  * 错误路径 HTTP status 走 ToHTTPStatus,body code 走 hecode.Code
  * 非 EcodeError 默认 code=ErrCodeUnknown(1099)
- middleware_test.go 追加 5 个 ResponseMiddleware 行为测试"
```

---

## Self-Review

**1. Spec coverage**
- 数据访问层(单实例 + 连接池 + 慢 SQL 日志)→ Task 1 ✓
- hecode → HTTP status 映射(业务错 200 / 系统错 5xx)→ Task 2 ✓
- 读写分离 / tx 跨层传递 / hollow-cli → 已按用户决策跳过

**2. Placeholder scan**
- 无 TBD / TODO / "实现 X" 等占位
- 每个 Step 都给了完整代码或具体命令

**3. Type consistency**
- `pkg/hgorm`: `Config` / `NewDB(ctx, dialector, cfg, log)` / `Ping(ctx, db)` / `NewZapGormLogger(log, slow, level)` 跨 Step 命名一致
- `pkg/hecode`: `ToHTTPStatus(err error) int` 与 ResponseMiddleware 调用点一致
- `internal/middleware`: `ResponseMiddleware` / `NewResponseMiddleware()` / `Response` 名称不变,外部调用兼容

**4. 风险点**
- `internal/middleware/middleware_test.go` 旧 test 没有写 ResponseMiddleware 测试,本任务只追加不删,如果改写时与文件现有 import 冲突,以现有为准并按需 merge
- `git rm reponse.go` 与 `git add response.go` 可能被 git 识别为 rename(取决于内容相似度),不影响合并
- 慢 SQL 日志的 `gormlogger.LogLevel` 接口签名以 v1.30 为准,若 GORM 升级到 v2 接口变化需同步
