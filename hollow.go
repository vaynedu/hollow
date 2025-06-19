package hollow

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/internal/config"
	"github.com/vaynedu/hollow/internal/middleware"
	"go.uber.org/zap"
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

type AppOption struct {
	Config      *config.Config
	Middlewares []gin.HandlerFunc // 中间件
	Logger      *zap.Logger       // 日志实例
}

func NewApp(opts AppOption) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		Ctx:    ctx,
		Cancel: cancel,
		Logger: opts.Logger,
		Engine: gin.New(),
	}

	// 初始化配置
	//cfg, err := config.NewConfig("conf.yaml")
	//if err != nil {
	//	return nil, err
	//}
	app.Config = opts.Config
	fmt.Println("app.Config", app.Config == nil)
	fmt.Printf("app.Config.LogConfig +%+v\n", app.Config.LogConfig)
	fmt.Printf("app.Config.Host +%+v\n", app.Config.GetString("host"))

	// 初始化日志
	//log, err := logger.InitLogger(cfg)
	//if err != nil {
	//	return nil, err
	//}
	app.Logger = opts.Logger

	// 导入默认的中间件
	defaultMiddlewares := middleware.RegisterDefaultMiddlewares(app.Logger)
	app.Middlewares = append(app.Middlewares, defaultMiddlewares...)
	// 依赖注入，让用户可以自定义中间件，通过外界传入，而不是在框架内部初始化，依赖耦合
	if len(opts.Middlewares) >= 0 {
		app.Middlewares = append(app.Middlewares, opts.Middlewares...)
	}
	app.Engine.Use(app.Middlewares...)
	// 初始化中间件， 不要直接代码依赖中写死，扩展性会很差
	//app.Middlewares = []gin.HandlerFunc{
	//	middleware.LoggingMiddleware(app.Logger),
	//	middleware.RecoveryMiddleware(app.Logger),
	//	middleware.ResponseMiddleware(),
	//}
	//app.Engine.Use(app.Middlewares...)

	return app, nil
}

// Start 启动服务
func (app *App) Start() error {
	app.Logger.Info("starting hollow server")
	addr := app.Config.GetString("host")
	if addr == "" {
		addr = ":8181"
	}
	return app.Engine.Run(addr)
}

func (app *App) End() {
	app.Logger.Info("stopping hollow server")
	app.Cancel()

	// todo 考虑优雅的关闭服务
}

func (app *App) AddMiddleware(middlewares ...gin.HandlerFunc) {
	app.Middlewares = append(app.Middlewares, middlewares...)
	app.Engine.Use(middlewares...)
}
