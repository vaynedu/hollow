package hollow_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	hollow "github.com/vaynedu/hollow"
)

func TestAppAcceptsExternalGinMiddleware(t *testing.T) {
	app := newTestApp(t)
	called := false
	app.Use(func(c *gin.Context) {
		called = true
		c.Next()
	})
	app.Engine.GET("/test", func(c *gin.Context) { c.Set("data", "ok") })

	rec := httptest.NewRecorder()
	app.Engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	if !called || rec.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, rec.Code)
	}
}

func TestAppDoesNotRegisterServiceHealthEndpoint(t *testing.T) {
	app := newTestApp(t)
	for _, route := range app.Engine.Routes() {
		if route.Path == "/-/health" {
			t.Fatalf("Hollow 不应注册服务级健康路由: %+v", route)
		}
	}
}

func newTestApp(t *testing.T) *hollow.App {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	content := []byte(`server:
  host: 127.0.0.1:8181
  shutdown_timeout: 5s
log:
  level: error
  output_mode: console
`)
	if err := os.WriteFile(filepath.Join(dir, "conf.yaml"), content, 0o600); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	app, err := hollow.NewApp(hollow.AppOption{ConfigPath: dir, ConfigName: "conf"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	return app
}
