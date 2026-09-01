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
