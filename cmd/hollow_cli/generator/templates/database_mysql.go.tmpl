package database

import (
	"context"

	"github.com/vaynedu/hollow/pkg/hgorm"
	"github.com/vaynedu/hollow/pkg/hlog"
	"gorm.io/gorm"
)

var mysqlDB *gorm.DB

// InitMySQL 根据 conf.yaml 初始化 MySQL；未启用时直接返回。
func InitMySQL() error {
	cfg, err := loadedConfig()
	if err != nil {
		return err
	}
	if !cfg.Database.MySQL.Enabled {
		return nil
	}

	db, err := hgorm.NewMySQL(context.Background(), cfg.Database.MySQL.Config)
	if err != nil {
		return err
	}
	mysqlDB = db
	hlog.L().Named("database.InitMySQL").Info("mysql initialized")
	return nil
}

// MySQL 返回已初始化的 GORM 实例；未启用时返回 nil。
func MySQL() *gorm.DB {
	return mysqlDB
}

func closeMySQL(ctx context.Context) error {
	if mysqlDB == nil {
		return nil
	}
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}
	hlog.FromContext(ctx).Named("database.Shutdown").Info("mysql closed")
	mysqlDB = nil
	return nil
}
