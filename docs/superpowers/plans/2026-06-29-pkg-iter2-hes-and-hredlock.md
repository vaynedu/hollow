# pkg Iter 2: 补 hes / hredlock 空壳包 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 hollow pkg 工程化迭代 2 落地:给两个空壳包 hes / hredlock 补真实封装,形态对齐 hredis(Config + NewClient + Ping)。同步升级 elasticsearch 客户端到 v8,并把 examples/es_demo 改写为基于新 hes 的演示。

**Architecture:**
- **hes**:基于 `github.com/elastic/go-elasticsearch/v8`,薄封装(Config + NewClient + Ping),不重造 ES 原生 API。调用方拿到 `*elasticsearch.Client` 后直接用 v8 的 `esapi.*` / `Search()` / `Index()` 等。
- **hredlock**:基于 `github.com/go-redsync/redsync/v4` + `github.com/redis/go-redis/v9`。Redlock 工厂接受外部传入的 `*redis.Client`(避免重复创建连接);Lock 返回 `*Mutex`,Mutex 上有 Unlock。隐藏 redsync 原生 API,但保留扩展空间。
- 升级 ES 依赖会让 examples/es_demo 用老 API 编译失败 —— 在 Task 1 一并改写它用新 hes,避免破坏 `go build ./...`。
- 测试无法连真 ES / Redis,只测构造与失败路径(nil/空配置),与 hredis 测试形态一致。

**Tech Stack:** Go 1.25 / github.com/elastic/go-elasticsearch/v8 / github.com/go-redsync/redsync/v4 / github.com/redis/go-redis/v9 / goconvey

## Global Constraints

- 包路径前缀 `github.com/vaynedu/hollow/pkg/...`
- pkg 内不得依赖 `internal/...`
- 中文 godoc 注释 + 中文 commit message
- Go 1.25
- 测试用 goconvey,不用 testify
- 不动 cmd/, internal/, example/, hollow.go
- 不动 pkg/ 其它子目录(只新加 hes/hredlock 内容)
- 唯一允许动的 examples/ 子目录:`examples/es_demo/`(因 ES 升级必须改)
- 任务结束前 `go build ./... && go test ./pkg/... -count=1` 全绿
- 新增依赖只允许 `github.com/elastic/go-elasticsearch/v8`(老 v0.0.0 替换);其它 redsync/go-redis 已在 go.mod
- 测试不依赖真实 ES / Redis 服务

## File Inventory

**修改:**
- `go.mod` / `go.sum`(`go get github.com/elastic/go-elasticsearch/v8` + `go mod tidy`)
- `examples/es_demo/main.go`(改写为用 hes + v8 API)

**新增:**
- `pkg/hes/es.go`(覆盖现有 1 行空 stub)
- `pkg/hes/es_test.go`
- `pkg/hes/example_test.go`
- `pkg/hes/doc.go`(本轮新增,Iter 1 跳过了空壳的 doc)
- `pkg/hredlock/redlock.go`
- `pkg/hredlock/redlock_test.go`(覆盖现有占位 test)
- `pkg/hredlock/example_test.go`
- `pkg/hredlock/doc.go`(本轮新增)

**不动:**
- `pkg/` 其它子目录(除上面两个空壳包)
- `examples/redis_pubsub/`、`examples/redlock_seckill/`(后者用的是 redsync 原生 API + go-redis/v9,自洽,本轮不强制改用 hredlock)
- 现有 `_test.go` 一律不动(只新建)

---

## Task 1: hes 薄封装 + ES v8 升级 + es_demo 重写

**Files:**
- Modify: `go.mod`, `go.sum`(升级 elasticsearch 到 v8)
- Modify: `pkg/hes/es.go`(从 1 行空 stub 重写为完整封装)
- Create: `pkg/hes/doc.go`
- Create: `pkg/hes/es_test.go`
- Create: `pkg/hes/example_test.go`
- Modify: `examples/es_demo/main.go`(改用 hes)

**Interfaces produces:**
- `type Config struct { Addresses []string; Username, Password, APIKey, CloudID string }`
- `var ErrEmptyAddresses = errors.New(...)`
- `func NewClient(ctx context.Context, cfg Config) (*elasticsearch.Client, error)` —— 构造 + Info() 探活
- `func Ping(ctx context.Context, client *elasticsearch.Client) error` —— 单独导出探活

**ES v8 API 关键点:**
- import path:`github.com/elastic/go-elasticsearch/v8`(client)、`github.com/elastic/go-elasticsearch/v8/esapi`(请求 API)
- `elasticsearch.NewClient(elasticsearch.Config{...})` 接口与 v0 兼容(字段名 Addresses/Username/Password/APIKey/CloudID 保持)
- 探活用 `client.Info(client.Info.WithContext(ctx))`(v8 API)
- 不支持顶层 `Timeout` 字段(若需要超时控制走 transport,本任务暂不暴露,调用方在 per-request 用 ctx 控制)

- [ ] **Step 1: 升级 elasticsearch 依赖**

```bash
cd /Users/nikki/go/src/hollow
go get github.com/elastic/go-elasticsearch/v8
go mod tidy
```

验证:
```bash
grep "elasticsearch" go.mod
```
Expected: 有 `github.com/elastic/go-elasticsearch/v8 vX.Y.Z`,老的 `v0.0.0` 自动消失

- [ ] **Step 2: 重写 `pkg/hes/es.go`**

```go
package hes

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

// Config hes 客户端配置
//
// Addresses 与 CloudID 至少二者填一:
//   - 普通自建 ES 集群:填 Addresses(如 ["http://127.0.0.1:9200"])
//   - Elastic Cloud:填 CloudID
//
// 认证方式三选一:Username+Password / APIKey / CloudID 内嵌
type Config struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	APIKey    string   `mapstructure:"api_key"`
	CloudID   string   `mapstructure:"cloud_id"`
}

// ErrEmptyAddresses Addresses 与 CloudID 都为空时返回
var ErrEmptyAddresses = errors.New("hes: addresses is required (unless cloud_id is set)")

// NewClient 构造 ES v8 客户端并 Info() 探活一次
//
// 不暴露 timeout 配置:调用方在每次请求时用 ctx 控制超时
func NewClient(ctx context.Context, cfg Config) (*elasticsearch.Client, error) {
	if len(cfg.Addresses) == 0 && cfg.CloudID == "" {
		return nil, ErrEmptyAddresses
	}
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		APIKey:    cfg.APIKey,
		CloudID:   cfg.CloudID,
	})
	if err != nil {
		return nil, fmt.Errorf("hes: new client: %w", err)
	}
	if err := Ping(ctx, client); err != nil {
		return nil, fmt.Errorf("hes: ping failed: %w", err)
	}
	return client, nil
}

// Ping 用 Info() API 探活,非 2xx 返回错误
func Ping(ctx context.Context, client *elasticsearch.Client) error {
	if client == nil {
		return errors.New("hes: nil client")
	}
	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("hes: ping status %s", res.Status())
	}
	return nil
}
```

- [ ] **Step 3: 创建 `pkg/hes/doc.go`**

```go
// Package hes 提供 github.com/elastic/go-elasticsearch/v8 的薄封装:
//   - Config 配置(支持 Addresses/CloudID + Basic Auth/APIKey)
//   - NewClient 构造时自动 Info() 探活
//   - Ping 单独导出供外部探活
//
// 不重造 ES 原生 API,获取到 *elasticsearch.Client 后直接调用 v8 的
// esapi.* / Search() / Index() 等方法。
package hes
```

- [ ] **Step 4: 创建 `pkg/hes/es_test.go`**

```go
package hes

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewClient(t *testing.T) {
	Convey("NewClient", t, func() {
		Convey("Addresses 与 CloudID 都为空返回 ErrEmptyAddresses", func() {
			_, err := NewClient(context.Background(), Config{})
			So(err, ShouldEqual, ErrEmptyAddresses)
		})

		Convey("无效 Addresses Ping 失败返错误", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 200*1000*1000) // 200ms
			defer cancel()
			_, err := NewClient(ctx, Config{
				Addresses: []string{"http://127.0.0.1:1"}, // 几乎一定连不上
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

注:测试里的 `200*1000*1000` 表示 200ms;如果实现时编译器抱怨需要 `time.Duration` 类型,在文件顶部 import `"time"` 然后改成 `200*time.Millisecond`。

- [ ] **Step 5: 创建 `pkg/hes/example_test.go`**

```go
package hes

import (
	"context"
	"fmt"
	"time"
)

// Example 不带 // Output:,需要可用的 ES 实例,仅展示调用形态
func ExampleNewClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := NewClient(ctx, Config{
		Addresses: []string{"http://127.0.0.1:9200"},
		Username:  "elastic",
		Password:  "changeme",
	})
	if err != nil {
		fmt.Println("connect failed:", err)
		return
	}

	// 拿到原生 *elasticsearch.Client 后,可直接用 v8 的 esapi.*
	res, _ := client.Info(client.Info.WithContext(ctx))
	defer res.Body.Close()
	fmt.Println(res.Status())
}
```

- [ ] **Step 6: 重写 `examples/es_demo/main.go` 用新 hes**

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/vaynedu/hollow/pkg/hes"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 通过 hes 创建 v8 客户端(NewClient 内部 Info() 验证连通)
	es, err := hes.NewClient(ctx, hes.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("创建 Elasticsearch 客户端失败: %s", err)
	}

	fmt.Println("成功连接到 Elasticsearch")

	// 拿到原生 *elasticsearch.Client 后,用 v8 esapi 调用
	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: "1111",
		Body:       bytes.NewReader([]byte(`{"name": "1111"}`)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es)
	if err != nil {
		log.Fatalf("插入数据失败: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Elasticsearch 返回错误: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Fatalf("解析响应失败: %s", err)
	}
	fmt.Printf("数据插入成功,ID: %s,版本: %d\n",
		result["_id"], int(result["_version"].(float64)))
}
```

- [ ] **Step 7: 验证编译 + 测试**

```bash
cd /Users/nikki/go/src/hollow
go build ./...
go test ./pkg/hes/... -v -count=1
```
Expected:
- build 全绿(包括 examples/es_demo)
- hes 测试 PASS(只测构造与失败路径,不依赖真实 ES)

- [ ] **Step 8: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add -A
git commit -m "feat(hes): 升级 ES 客户端到 v8 并补薄封装

- go.mod 升级 elasticsearch v0.0.0 → v8.x
- pkg/hes 实现 Config / NewClient(自带 Info 探活) / Ping
- 新增 doc.go(中文 godoc 入口)与 example_test.go
- examples/es_demo 改写为通过 hes 创建客户端,演示 esapi 原生用法
- 测试覆盖 nil/空配置/无效地址,不依赖真实 ES"
```

---

## Task 2: hredlock Lock/Unlock 二件套

**Files:**
- Modify: `pkg/hredlock/redlock.go`(原文件不存在;新建)—— 实际是 Create
- Modify: `pkg/hredlock/redlock_test.go`(覆盖现有占位 placeholder test)
- Create: `pkg/hredlock/doc.go`
- Create: `pkg/hredlock/example_test.go`

**Interfaces produces:**
- `type Redlock struct { ... }` —— 锁工厂
- `type Mutex struct { ... }` —— 一把分布式锁
- `var ErrNilClient = errors.New(...)`
- `var ErrNilMutex = errors.New(...)`
- `var ErrUnlockFailed = errors.New(...)`
- `func NewRedlock(client *redis.Client) (*Redlock, error)` —— 接受外部 redis 客户端
- `func (r *Redlock) Lock(ctx context.Context, key string, expiry time.Duration) (*Mutex, error)` —— 加锁(阻塞,有 redsync 内部重试)
- `func (m *Mutex) Unlock(ctx context.Context) error` —— 释放锁

**设计说明:**
- 包装 `*redsync.Mutex` 隐藏 redsync 命名,但保留 Mutex 概念,便于以后扩展(如续期、TTL 查询)
- 锁 expiry 必须由调用方显式传入(避免框架强加默认值导致业务不感知)
- TryLock(不阻塞)按用户决策**不实现**;有需要时再加

- [ ] **Step 1: 创建 `pkg/hredlock/redlock.go`**

```go
package hredlock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// 包级错误
var (
	ErrNilClient    = errors.New("hredlock: nil redis client")
	ErrNilMutex     = errors.New("hredlock: nil mutex")
	ErrUnlockFailed = errors.New("hredlock: unlock failed (lock may have expired or be held by another)")
)

// Redlock 分布式锁工厂,内部包一个 redsync 实例
type Redlock struct {
	rs *redsync.Redsync
}

// NewRedlock 用业务方已建立的 *redis.Client 构造分布式锁工厂
// 不重复建立 redis 连接;调用方负责管理 client 的生命周期
func NewRedlock(client *redis.Client) (*Redlock, error) {
	if client == nil {
		return nil, ErrNilClient
	}
	pool := goredis.NewPool(client)
	return &Redlock{rs: redsync.New(pool)}, nil
}

// Mutex 一把分布式锁,wraps *redsync.Mutex
type Mutex struct {
	m *redsync.Mutex
}

// Lock 阻塞加锁,redsync 内部有重试与延迟
//
// expiry: 锁的自动过期时间(防止持锁进程崩溃后死锁)
// 业务应根据临界区耗时设定,建议 ≥ 临界区 P99 耗时 × 2,
// 同时配合在临界区结束后显式调用 Unlock
func (r *Redlock) Lock(ctx context.Context, key string, expiry time.Duration) (*Mutex, error) {
	if r == nil || r.rs == nil {
		return nil, ErrNilClient
	}
	m := r.rs.NewMutex(key, redsync.WithExpiry(expiry))
	if err := m.LockContext(ctx); err != nil {
		return nil, fmt.Errorf("hredlock: lock %q: %w", key, err)
	}
	return &Mutex{m: m}, nil
}

// Unlock 释放锁
//
// 锁已被其它持有者抢占或自动过期时返回 ErrUnlockFailed
// 业务应监控该错误以发现临界区超时问题
func (m *Mutex) Unlock(ctx context.Context) error {
	if m == nil || m.m == nil {
		return ErrNilMutex
	}
	ok, err := m.m.UnlockContext(ctx)
	if err != nil {
		return fmt.Errorf("hredlock: unlock: %w", err)
	}
	if !ok {
		return ErrUnlockFailed
	}
	return nil
}
```

- [ ] **Step 2: 创建 `pkg/hredlock/doc.go`**

```go
// Package hredlock 提供基于 github.com/go-redsync/redsync/v4 的分布式锁封装。
//
// 设计:
//   - Redlock 工厂接受业务已建立的 *redis.Client,不重复创建连接
//   - Lock 返回 *Mutex,*Mutex.Unlock 释放(典型 defer 调用)
//   - 锁 expiry 必须显式传入,避免框架默认值掩盖业务超时风险
//
// 不暴露 redsync 原生选项(WithTries / WithRetryDelay 等);
// 后续按需扩展。
package hredlock
```

- [ ] **Step 3: 覆盖 `pkg/hredlock/redlock_test.go`**

```go
package hredlock

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/redis/go-redis/v9"
)

func TestNewRedlock(t *testing.T) {
	Convey("NewRedlock", t, func() {
		Convey("nil client 返回 ErrNilClient", func() {
			_, err := NewRedlock(nil)
			So(err, ShouldEqual, ErrNilClient)
		})

		Convey("非 nil client 构造成功", func() {
			client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
			defer client.Close()
			rl, err := NewRedlock(client)
			So(err, ShouldBeNil)
			So(rl, ShouldNotBeNil)
		})
	})
}

func TestLock_NilRedlock(t *testing.T) {
	Convey("nil Redlock.Lock 返回 ErrNilClient", t, func() {
		var rl *Redlock
		_, err := rl.Lock(context.Background(), "k", time.Second)
		So(err, ShouldEqual, ErrNilClient)
	})
}

func TestLock_UnreachableRedis(t *testing.T) {
	Convey("不可达 Redis 加锁失败", t, func() {
		client := redis.NewClient(&redis.Options{
			Addr:        "127.0.0.1:1",
			DialTimeout: 100 * time.Millisecond,
		})
		defer client.Close()
		rl, _ := NewRedlock(client)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := rl.Lock(ctx, "test-key", time.Second)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "lock")
	})
}

func TestUnlock_NilMutex(t *testing.T) {
	Convey("nil Mutex.Unlock", t, func() {
		Convey("nil receiver 返回 ErrNilMutex", func() {
			var m *Mutex
			err := m.Unlock(context.Background())
			So(err, ShouldEqual, ErrNilMutex)
		})
		Convey("nil 内部 m 返回 ErrNilMutex", func() {
			m := &Mutex{}
			err := m.Unlock(context.Background())
			So(err, ShouldEqual, ErrNilMutex)
		})
	})
}
```

- [ ] **Step 4: 创建 `pkg/hredlock/example_test.go`**

```go
package hredlock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Example 不带 // Output:,需要可用的 Redis 实例,仅展示调用形态
func ExampleRedlock_Lock() {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer client.Close()

	rl, err := NewRedlock(client)
	if err != nil {
		fmt.Println("init failed:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 加锁,过期时间 5s
	mu, err := rl.Lock(ctx, "order:1001", 5*time.Second)
	if err != nil {
		fmt.Println("lock failed:", err)
		return
	}
	defer func() {
		if err := mu.Unlock(context.Background()); err != nil {
			fmt.Println("unlock failed:", err)
		}
	}()

	// 临界区...
}
```

- [ ] **Step 5: 验证编译 + 测试**

```bash
cd /Users/nikki/go/src/hollow
go build ./...
go test ./pkg/hredlock/... -v -count=1
```
Expected: 全 PASS,无新增依赖(redsync/go-redis 已在 go.mod)

- [ ] **Step 6: 提交**

```bash
cd /Users/nikki/go/src/hollow
git add pkg/hredlock
git commit -m "feat(hredlock): 基于 redsync/v4 实现分布式锁 Lock/Unlock 二件套

- NewRedlock 接受外部 *redis.Client(避免重复建立连接)
- Lock 阻塞加锁,显式 expiry 防死锁
- Mutex.Unlock 释放,Lock 已超时/被抢时返回 ErrUnlockFailed
- 不暴露 redsync 原生选项,后续按需扩展
- 测试覆盖 nil 边界 + 不可达 Redis 失败路径"
```

---

## Self-Review

**1. Spec coverage**
- Iter 2 目标 = hes/hredlock 补真实封装:Task 1 hes ✓,Task 2 hredlock ✓
- ES v8 升级 + es_demo 重写:在 Task 1 范围内
- Iter 3 / 4 不在本 plan 范围

**2. Placeholder scan**
- 无 TBD / TODO 占位
- 每个 Step 都给完整代码或具体命令

**3. Type consistency**
- hes:`Config` / `NewClient(ctx, cfg)` / `Ping(ctx, client)` / `ErrEmptyAddresses` —— 跨 Step 命名一致
- hredlock:`Redlock` / `Mutex` / `NewRedlock(client)` / `Lock(ctx, key, expiry)` / `Unlock(ctx)` / `ErrNilClient` / `ErrNilMutex` / `ErrUnlockFailed` —— 跨 Step 一致
- examples/es_demo:用 `hes.NewClient` 而非旧 `elasticsearch.NewClient`,函数签名与 Task 1 Step 2 对齐

**4. 风险点**
- ES v8 客户端 `elasticsearch.Config` 字段名与 v0 保持一致(Addresses/Username/Password/APIKey/CloudID)—— 这是 ES Go 客户端的稳定 API,实测应该兼容
- `client.Info(client.Info.WithContext(ctx))` 是 v8 的标准 API,若实施时找不到 WithContext 方法名,降级为 `client.Info(esapi.WithContext(ctx))` 或最简 `client.Info()` 然后用 ctx 控制 transport
- `go-redsync/redsync/v4/redis/goredis/v9` 适配器要 redsync v4.7+,当前 v4.13.0 满足
- 测试用 `127.0.0.1:1` 不可达地址依赖 OS 行为(本地 1 端口通常 RST)—— Linux/macOS 表现稳定;若 CI 在某些容器内 1 端口被代理,测试可能 hang,可改 `127.0.0.1:65535` 或 `127.0.0.1:0` 之类(后者不合法可立即 RST)
