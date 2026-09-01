package hollow

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Run 启动 HTTP Server，并在收到退出信号后优雅关闭应用。
func (app *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return app.run(ctx, net.Listen)
}

func (app *App) run(ctx context.Context, listen func(string, string) (net.Listener, error)) error {
	if err := app.runStartupHooks(); err != nil {
		return errors.Join(err, app.shutdown(nil))
	}

	listener, err := listen("tcp", app.config.Server.Host)
	if err != nil {
		return errors.Join(err, app.shutdown(nil))
	}

	server := &http.Server{Handler: app.Engine}
	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- server.Serve(listener)
	}()

	var serveErr error
	serveStopped := false
	select {
	case <-ctx.Done():
	case serveErr = <-serveErrors:
		serveStopped = true
	}

	shutdownErr := app.shutdown(server)
	if !serveStopped {
		serveErr = <-serveErrors
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}

	return errors.Join(serveErr, shutdownErr)
}

func (app *App) shutdown(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), app.config.Server.ShutdownTimeout)
	defer cancel()

	var serverErr error
	if server != nil {
		serverErr = server.Shutdown(ctx)
	}
	hookErr := app.runShutdownHooks(ctx)
	loggerErr := app.Logger.Sync()
	if errors.Is(loggerErr, syscall.EINVAL) || errors.Is(loggerErr, syscall.ENOTTY) {
		loggerErr = nil
	}

	return errors.Join(serverErr, hookErr, loggerErr)
}
