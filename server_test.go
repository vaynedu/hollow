package hollow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRunStartupFailureStillRunsAllShutdownHooks(t *testing.T) {
	startupErr := errors.New("startup failed")
	shutdownErr := errors.New("shutdown failed")
	app := newUnitApp(t)
	var order []string
	app.Startup(func() error { return startupErr })
	app.Shutdown(
		func(context.Context) error {
			order = append(order, "first")
			return shutdownErr
		},
		func(context.Context) error {
			order = append(order, "second")
			return nil
		},
	)

	listenCalled := false
	err := app.run(context.Background(), func(string, string) (net.Listener, error) {
		listenCalled = true
		return nil, nil
	})

	if !errors.Is(err, startupErr) || !errors.Is(err, shutdownErr) {
		t.Fatalf("error=%v want joined errors %v and %v", err, startupErr, shutdownErr)
	}
	if listenCalled {
		t.Fatal("启动 Hook 失败后不应监听端口")
	}
	if want := []string{"second", "first"}; !slices.Equal(order, want) {
		t.Fatalf("shutdown order=%v want=%v", order, want)
	}
}

func TestRunReturnsListenErrorAndCleansUp(t *testing.T) {
	listenErr := errors.New("listen failed")
	cleaned := false
	app := newUnitApp(t)
	app.Shutdown(func(context.Context) error {
		cleaned = true
		return nil
	})

	err := app.run(context.Background(), func(string, string) (net.Listener, error) {
		return nil, listenErr
	})

	if !errors.Is(err, listenErr) || !cleaned {
		t.Fatalf("error=%v cleaned=%v", err, cleaned)
	}
}

func TestRunStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	app := newUnitApp(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	var shutdownContextErr error
	app.Shutdown(func(ctx context.Context) error {
		shutdownContextErr = ctx.Err()
		return nil
	})
	cancel()
	err = app.run(ctx, func(string, string) (net.Listener, error) { return ln, nil })

	if err != nil {
		t.Fatalf("run error=%v", err)
	}
	if shutdownContextErr != nil {
		t.Fatalf("shutdown context error=%v, want context.Background 派生的新上下文", shutdownContextErr)
	}
}

func TestRunServesHealthEndpointUntilCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	app := newUnitApp(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	runErr := make(chan error, 1)
	go func() {
		runErr <- app.run(ctx, func(string, string) (net.Listener, error) { return ln, nil })
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + ln.Addr().String() + "/-/health")
	if err != nil {
		cancel()
		t.Fatalf("请求健康检查失败: %v", err)
	}
	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		cancel()
		t.Fatalf("读取健康检查响应失败: %v", readErr)
	}
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"status":"ok"`) {
		cancel()
		t.Fatalf("status=%d body=%s", resp.StatusCode, body)
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("run error=%v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("取消上下文后 Run 未退出")
	}
}

func TestRunJoinsServeHookAndLoggerErrors(t *testing.T) {
	serveErr := errors.New("serve failed")
	shutdownErr := errors.New("hook shutdown failed")
	loggerErr := errors.New("logger sync failed")
	app := newUnitApp(t)
	app.Logger = newSyncErrorLogger(loggerErr)
	app.Shutdown(func(context.Context) error { return shutdownErr })

	err := app.run(context.Background(), func(string, string) (net.Listener, error) {
		return &acceptErrorListener{err: serveErr}, nil
	})

	for _, want := range []error{serveErr, shutdownErr, loggerErr} {
		if !errors.Is(err, want) {
			t.Fatalf("error=%v want joined error %v", err, want)
		}
	}
}

func TestRunJoinsServerShutdownHookAndLoggerErrors(t *testing.T) {
	serverShutdownErr := errors.New("server shutdown failed")
	hookShutdownErr := errors.New("hook shutdown failed")
	loggerErr := errors.New("logger sync failed")
	ctx, cancel := context.WithCancel(context.Background())
	app := newUnitApp(t)
	app.Logger = newSyncErrorLogger(loggerErr)
	app.Shutdown(func(context.Context) error { return hookShutdownErr })
	ln := newBlockingListener(serverShutdownErr)

	runErr := make(chan error, 1)
	go func() {
		runErr <- app.run(ctx, func(string, string) (net.Listener, error) { return ln, nil })
	}()
	select {
	case <-ln.acceptStarted:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP Server 未开始接受连接")
	}

	select {
	case err := <-runErr:
		for _, want := range []error{serverShutdownErr, hookShutdownErr, loggerErr} {
			if !errors.Is(err, want) {
				t.Fatalf("error=%v want joined error %v", err, want)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("取消上下文后 Run 未退出")
	}
}

func TestRunIgnoresCommonTerminalLoggerSyncErrors(t *testing.T) {
	listenErr := errors.New("listen failed")
	tests := []struct {
		name    string
		syncErr error
	}{
		{name: "EINVAL", syncErr: fmt.Errorf("sync stdout: %w", syscall.EINVAL)},
		{name: "ENOTTY", syncErr: fmt.Errorf("sync stderr: %w", syscall.ENOTTY)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newUnitApp(t)
			app.Logger = newSyncErrorLogger(tt.syncErr)
			err := app.run(context.Background(), func(string, string) (net.Listener, error) {
				return nil, listenErr
			})

			if !errors.Is(err, listenErr) {
				t.Fatalf("error=%v want=%v", err, listenErr)
			}
			if errors.Is(err, tt.syncErr) {
				t.Fatalf("error=%v should ignore logger sync error %v", err, tt.syncErr)
			}
		})
	}
}

func newUnitApp(t *testing.T) *App {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	content := []byte(`server:
  host: 127.0.0.1:0
  shutdown_timeout: 1s
log:
  level: error
  output_mode: console
`)
	if err := os.WriteFile(filepath.Join(dir, "conf.yaml"), content, 0o600); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	app, err := NewApp(AppOption{ConfigPath: dir, ConfigName: "conf"})
	if err != nil {
		t.Fatalf("NewApp() error=%v", err)
	}
	app.Logger = newSyncErrorLogger(nil)
	return app
}

type syncErrorWriter struct {
	err error
}

func (w *syncErrorWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (w *syncErrorWriter) Sync() error {
	return w.err
}

func newSyncErrorLogger(err error) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		&syncErrorWriter{err: err},
		zap.DebugLevel,
	)
	return zap.New(core)
}

type acceptErrorListener struct {
	err error
}

func (l *acceptErrorListener) Accept() (net.Conn, error) { return nil, l.err }
func (l *acceptErrorListener) Close() error              { return nil }
func (l *acceptErrorListener) Addr() net.Addr            { return testAddr("accept-error") }

type blockingListener struct {
	acceptStarted chan struct{}
	closed        chan struct{}
	acceptOnce    sync.Once
	closeOnce     sync.Once
	closeErr      error
}

func newBlockingListener(closeErr error) *blockingListener {
	return &blockingListener{
		acceptStarted: make(chan struct{}),
		closed:        make(chan struct{}),
		closeErr:      closeErr,
	}
}

func (l *blockingListener) Accept() (net.Conn, error) {
	l.acceptOnce.Do(func() { close(l.acceptStarted) })
	<-l.closed
	return nil, net.ErrClosed
}

func (l *blockingListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return l.closeErr
}

func (l *blockingListener) Addr() net.Addr { return testAddr("blocking") }

type testAddr string

func (a testAddr) Network() string { return "tcp" }
func (a testAddr) String() string  { return string(a) }
