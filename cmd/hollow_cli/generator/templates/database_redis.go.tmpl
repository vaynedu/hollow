package database

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/vaynedu/hollow/pkg/hlog"
	"github.com/vaynedu/hollow/pkg/hredis"
)

var redisClient *redis.Client

// InitRedis 根据 conf.yaml 初始化 Redis；未启用时直接返回。
func InitRedis() error {
	cfg, err := loadedConfig()
	if err != nil {
		return err
	}
	if !cfg.Database.Redis.Enabled {
		return nil
	}

	client, err := hredis.NewClient(context.Background(), cfg.Database.Redis.Config)
	if err != nil {
		return err
	}
	redisClient = client
	hlog.L().Named("database.InitRedis").Info("redis initialized")
	return nil
}

// Redis 返回已初始化的 Redis 客户端；未启用时返回 nil。
func Redis() *redis.Client {
	return redisClient
}

func closeRedis(ctx context.Context) error {
	if redisClient == nil {
		return nil
	}
	if err := redisClient.Close(); err != nil {
		return err
	}
	hlog.FromContext(ctx).Named("database.Shutdown").Info("redis closed")
	redisClient = nil
	return nil
}
