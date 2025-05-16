package hollow

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hollow/internal/config"
	"hollow/internal/logger"
	"hollow/internal/middleware"
)

// App 框架核心结构体
type App struct {
	Ctx         context.Context    // 全局上下文
	Cancel      context.CancelFunc // 取消上下文
	Config      *config.Config     // 配置管理器
	Logger      *zap.Logger        // 日志实例
	Engine      *gin.Engine        // gin引擎实例
	Middlewares []gin.HandlerFunc  // 中间件
}

func NewApp(cfgPath string) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		Ctx:    ctx,
		Cancel: cancel,
		Logger: zap.NewNop(),
		Engine: gin.New(),
	}

	// 初始化配置
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	app.Config = cfg

	// 初始化日志
	log, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, err
	}
	app.Logger = log

	// 初始化中间件
	app.Middlewares = []gin.HandlerFunc{
		middleware.LoggingMiddleware(app.Logger),
		middleware.RecoveryMiddleware(app.Logger),
		middleware.ResponseMiddleware(),
	}
	app.Engine.Use(app.Middlewares...)

	return app, nil
}

func (app *App) Start() error {
	app.Logger.Info("starting hollow server")
	return nil
}

func (app *App) End() {
	app.Logger.Info("stopping hollow server")
	app.Cancel()
}
