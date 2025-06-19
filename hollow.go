package hollow

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"slices"

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
	fmt.Printf("app.Config.LogConfig +%+v\n", app.Config.Log)
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
func (app *App) Start() {
	go func() {
		app.Logger.Info("starting hollow server")
		addr := app.Config.GetString("host")
		if addr == "" {
			addr = ":8181"
		}
		if err := app.Engine.Run(addr); err != nil {
			app.Logger.Fatal("failed to start server", zap.Error(err))
		}
	}()
}

func (app *App) End() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("stopping hollow server")
	app.Cancel()

	// todo 考虑优雅的关闭服务
}

func (app *App) AddMiddleware(middlewares ...gin.HandlerFunc) {
	app.Middlewares = append(app.Middlewares, middlewares...)
	app.Engine.Use(middlewares...)
}

func (app *App) AddRoute(method, path string, handlerFunc gin.HandlerFunc) {
	app.Engine.Handle(method, path, handlerFunc)
}

func (app *App) RemoveMiddleware(middlewares ...gin.HandlerFunc) {
	// 用户可以指定不使用某个中间件
	// WARNING 感觉这段代码有风险， 因为go的函数不能比较，有可能地址一样，但是行为不同，比如闭包
	// 但是对于中间件，一般都是在初始化的时候就确定了，不会在运行时动态改变，所以认为是安全的
	// 如果针对函数一定要比较，可以封装一个标识符
	for _, middleware := range middlewares {
		sf1 := reflect.ValueOf(middleware)
		for i := 0; i < len(app.Middlewares); i++ {
			sf2 := reflect.ValueOf(app.Middlewares[i])
			if sf1.Pointer() == sf2.Pointer() {
				app.Middlewares = slices.Delete(app.Middlewares, i, i+1)
				i-- // 调整索引
			}
		}
	}
	app.Engine.Use(middlewares...)
}
