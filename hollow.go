package hollow

import (
	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/internal/config"
	"github.com/vaynedu/hollow/internal/logger"
	"github.com/vaynedu/hollow/internal/middleware"
	"go.uber.org/zap"
)

// App 框架核心结构体
type App struct {
	Engine        *gin.Engine
	Logger        *zap.Logger
	config        *config.Config
	startupHooks  []StartupFunc
	shutdownHooks []ShutdownFunc
}

type AppOption struct {
	ConfigPath string
	ConfigName string
}

func NewApp(opts AppOption) (*App, error) {
	// 处理默认配置路径和名称
	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = "."
	}
	configName := opts.ConfigName
	if configName == "" {
		configName = "conf"
	}

	// 初始化配置
	cfg, err := config.NewConfig(configPath, configName)
	if err != nil {
		return nil, err
	}

	// 初始化日志
	log, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, err
	}

	app := &App{
		Engine: gin.New(),
		Logger: log,
		config: cfg,
	}
	app.Engine.Use(middleware.RegisterDefaultMiddlewares(app.Logger)...)

	return app, nil
}

func (app *App) AddRoute(method, path string, handlerFunc gin.HandlerFunc) {
	app.Engine.Handle(method, path, handlerFunc)
}

// Group 创建路由组
func (app *App) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return app.Engine.Group(relativePath, handlers...)
}

// Use 注册 Gin 原生中间件。
func (app *App) Use(handlers ...gin.HandlerFunc) {
	app.Engine.Use(handlers...)
}
