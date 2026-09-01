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
