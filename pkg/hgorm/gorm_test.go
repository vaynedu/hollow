package hgorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vaynedu/hollow/pkg/hlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
)

func TestNewDB(t *testing.T) {
	Convey("NewDB 初始化 GORM 和连接池", t, func() {
		Convey("拒绝 nil Dialector", func() {
			_, err := NewDB(context.Background(), nil, Config{})

			So(errors.Is(err, ErrNilDialector), ShouldBeTrue)
		})

		Convey("设置默认连接池并执行 Ping", func() {
			sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			So(err, ShouldBeNil)
			defer sqlDB.Close()
			mock.ExpectPing()

			dialector := mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			})
			db, err := NewDB(context.Background(), dialector, Config{})

			So(err, ShouldBeNil)
			So(db, ShouldNotBeNil)
			So(sqlDB.Stats().MaxOpenConnections, ShouldEqual, 50)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

func TestNewMySQLAndPingValidation(t *testing.T) {
	Convey("MySQL 配置和 Ping 校验", t, func() {
		_, err := NewMySQL(context.Background(), Config{})
		So(errors.Is(err, ErrEmptyDSN), ShouldBeTrue)

		err = Ping(context.Background(), nil)
		So(errors.Is(err, ErrNilDB), ShouldBeTrue)
	})
}

func TestGORMLoggerUsesContextLogger(t *testing.T) {
	Convey("GORM Logger 继承 Context 日志字段", t, func() {
		core, logs := observer.New(zapcore.DebugLevel)
		ctxLogger := zap.New(core).With(hlog.String("request_id", "request-1"))
		ctx := hlog.NewContext(context.Background(), ctxLogger)
		gormLogger := newGORMLogger(10*time.Millisecond, logger.Warn)

		gormLogger.Trace(ctx, time.Now().Add(-20*time.Millisecond), func() (string, int64) {
			return "SELECT 1", 1
		}, nil)

		So(logs.Len(), ShouldEqual, 1)
		entry := logs.All()[0]
		So(entry.Message, ShouldEqual, "gorm slow sql")
		So(entry.ContextMap()["request_id"], ShouldEqual, "request-1")
	})
}
