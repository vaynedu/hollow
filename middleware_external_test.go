package hollow_test

import (
	"encoding/json"
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

func TestHealthEndpoint(t *testing.T) {
	app := newTestApp(t)
	rec := httptest.NewRecorder()
	app.Engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/-/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析健康检查响应失败: %v", err)
	}
	if response.Data.Status != "ok" {
		t.Fatalf("health status=%q", response.Data.Status)
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
