package hgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// ErrEmptyDSN 表示 MySQL DSN 未配置。
	ErrEmptyDSN = errors.New("hgorm: dsn is required")
	// ErrNilDialector 表示未提供 GORM 数据库驱动。
	ErrNilDialector = errors.New("hgorm: dialector is required")
	// ErrNilDB 表示数据库实例为空。
	ErrNilDB = errors.New("hgorm: nil db")
)

// Config 定义 MySQL、连接池和 SQL 日志配置。
type Config struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
	LogLevel        string        `mapstructure:"log_level"`
}

// NewMySQL 创建 MySQL GORM 实例。
func NewMySQL(ctx context.Context, cfg Config) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, ErrEmptyDSN
	}
	return NewDB(ctx, mysql.Open(cfg.DSN), cfg)
}

// NewDB 使用指定 Dialector 创建 GORM 实例，设置连接池并执行 Ping。
func NewDB(ctx context.Context, dialector gorm.Dialector, cfg Config) (*gorm.DB, error) {
	if dialector == nil {
		return nil, ErrNilDialector
	}
	applyDefaults(&cfg)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:               newGORMLogger(cfg.SlowThreshold, parseLogLevel(cfg.LogLevel)),
		DisableAutomaticPing: true,
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

// Ping 检查 GORM 底层数据库连接。
func Ping(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return ErrNilDB
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
