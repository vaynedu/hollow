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
