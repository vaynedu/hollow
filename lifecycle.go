package hollow

import (
	"context"
	"errors"
)

// StartupFunc 定义应用启动 Hook。
type StartupFunc func() error

// ShutdownFunc 定义应用关闭 Hook。
type ShutdownFunc func(context.Context) error

// Startup 注册应用启动 Hook。
func (app *App) Startup(hooks ...StartupFunc) *App {
	app.startupHooks = append(app.startupHooks, hooks...)
	return app
}

// Shutdown 注册应用关闭 Hook。
func (app *App) Shutdown(hooks ...ShutdownFunc) *App {
	app.shutdownHooks = append(app.shutdownHooks, hooks...)
	return app
}

func (app *App) runStartupHooks() error {
	for _, hook := range app.startupHooks {
		if err := hook(); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) runShutdownHooks(ctx context.Context) error {
	var errs []error
	for i := len(app.shutdownHooks) - 1; i >= 0; i-- {
		if err := app.shutdownHooks[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
