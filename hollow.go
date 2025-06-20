package hollow

import (
	"context"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/internal/config"
	"github.com/vaynedu/hollow/internal/logger"
	"github.com/vaynedu/hollow/internal/middleware"
	"go.uber.org/zap"
)

// App 框架核心结构体
type App struct {
	Ctx         context.Context         // 全局上下文
	Cancel      context.CancelFunc      // 取消上下文
	Config      *config.Config          // 配置管理器
	Logger      *zap.Logger             // 日志实例
	Engine      *gin.Engine             // gin引擎实例
	Middlewares []middleware.Middleware // 中间件
}

type AppOption struct {
	ConfigPath        string                  // 配置文件路径
	ConfigName        string                  // 配置文件名
	AddMiddlewares    []middleware.Middleware // 增加中间件
	RemoveMiddlewares []middleware.Middleware // 移除中间件
}

func NewApp(opts AppOption) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		Ctx:    ctx,
		Cancel: cancel,
		Engine: gin.New(),
	}

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
	app.Config = cfg

	// 初始化日志
	log, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, err
	}
	app.Logger = log

	// 导入默认的中间件
	defaultMiddlewares := middleware.RegisterDefaultMiddlewares(app.Logger)
	app.AddMiddleware(defaultMiddlewares...)
	// 依赖注入，让用户可以自定义中间件
	if len(opts.AddMiddlewares) > 0 {
		app.AddMiddleware(opts.AddMiddlewares...)
	}
	// 移除用户指定的中间件
	if len(opts.RemoveMiddlewares) > 0 {
		app.RemoveMiddleware(opts.RemoveMiddlewares...)
	}
	app.UseMiddleware(app.Middlewares...)

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
	app.Logger.Sync() // 确保日志正确刷新
}

func (app *App) AddRoute(method, path string, handlerFunc gin.HandlerFunc) {
	app.Engine.Handle(method, path, handlerFunc)
}

func (app *App) UseMiddleware(middlewares ...middleware.Middleware) {
	// 将 middleware.Middleware 类型的切片转换为 gin.HandlerFunc 类型的切片
	handlerFuncs := make([]gin.HandlerFunc, 0)
	for _, m := range middlewares {
		handlerFuncs = append(handlerFuncs, m.HandlerFunc())
	}
	app.Engine.Use(handlerFuncs...)
}

// // AddMiddleware 增加中间件
// func (app *App) AddMiddleware(middlewares ...gin.HandlerFunc) {
// 	app.Middlewares = append(app.Middlewares, middlewares...)
// }

// AddMiddleware 增加中间件
func (app *App) AddMiddleware(middlewares ...middleware.Middleware) {
	// 新增的时候考虑去重，如果重复就跳过
	for _, m := range middlewares {
		found := false
		for _, existing := range app.Middlewares {
			if m.Identifier() == existing.Identifier() {
				app.Logger.Warn("middleware already exists", zap.String("identifier", m.Identifier()))
				found = true
				break
			}
		}
		if !found {
			app.Middlewares = append(app.Middlewares, m)
		}
	}
}

func (app *App) RemoveMiddleware(middlewares ...middleware.Middleware) {
	// 移除的时候考虑一下不存在， 如果不存在就跳过
	for _, middleware := range middlewares {
		for i := 0; i < len(app.Middlewares); i++ {
			if middleware.Identifier() == app.Middlewares[i].Identifier() {
				app.Middlewares = slices.Delete(app.Middlewares, i, i+1)
				i-- // 调整索引
			}
		}
	}
}

// // RemoveMiddleware 移除中间件
// func (app *App) RemoveMiddleware(middlewares ...gin.HandlerFunc) {
// 	// 用户可以指定不使用某个中间件
// 	// WARNING 感觉这段代码有风险， 因为go的函数不能比较，有可能地址一样，但是行为不同，比如闭包
// 	// 但是对于中间件，一般都是在初始化的时候就确定了，不会在运行时动态改变，所以认为是安全的
// 	// 如果针对函数一定要比较，可以封装一个标识符

// 	// 每次调用都会返回 比如 ResponseMiddleware()都会返回一个新的匿名函数实例, 因此下面做法就是失败，而且风险大
// 	for _, middleware := range middlewares {
// 		sf1 := reflect.ValueOf(middleware)
// 		for i := 0; i < len(app.Middlewares); i++ {
// 			sf2 := reflect.ValueOf(app.Middlewares[i])
// 			if sf1.Pointer() == sf2.Pointer() {
// 				app.Middlewares = slices.Delete(app.Middlewares, i, i+1)
// 				i-- // 调整索引
// 			}
// 		}
// 	}
// }
