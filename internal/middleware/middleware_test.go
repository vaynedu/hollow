package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hecode"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type responseEnvelope struct {
	Code      int             `json:"code"`
	Msg       string          `json:"msg"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

func TestRequestIDMiddlewarePropagatesOneID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewRequestIDMiddleware())
	router.GET("/", func(c *gin.Context) {
		ginID, _ := c.Get(RequestIDKey)
		ctxID := c.Request.Context().Value(requestIDContextKey{})
		c.String(http.StatusOK, "%v/%v", ginID, ctxID)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "incoming-id")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") != "incoming-id" || rec.Body.String() != "incoming-id/incoming-id" {
		t.Fatalf("header=%q body=%q", rec.Header().Get("X-Request-ID"), rec.Body.String())
	}
}

func TestMetricsMiddlewareCanBeRegisteredDirectly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core, logs := observer.New(zapcore.InfoLevel)
	router := gin.New()
	router.Use(NewMetricsMiddleware(zap.New(core)))
	router.GET("/metrics", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("metrics logs=%d", len(entries))
	}
	fields := entries[0].ContextMap()
	if entries[0].Message != "请求耗时" || fields["method"] != http.MethodGet || fields["path"] != "/metrics" {
		t.Fatalf("metrics log=%+v fields=%v", entries[0], fields)
	}
}

func TestDefaultMiddlewarePipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core, logs := observer.New(zapcore.DebugLevel)
	router := gin.New()
	router.Use(RegisterDefaultMiddlewares(zap.New(core))...)
	router.GET("/success", func(c *gin.Context) { c.Set("data", gin.H{"ok": true}) })
	router.GET("/forbidden", func(c *gin.Context) { _ = c.Error(hecode.ErrForbidden) })
	router.GET("/plain", func(c *gin.Context) { _ = c.Error(errors.New("secret")) })
	router.GET("/panic", func(c *gin.Context) { panic("secret panic") })

	success := assertEnvelope(t, router, "/success", http.StatusOK, 0, "success")
	assertJSONData(t, success.Data, map[string]any{"ok": true})
	assertEnvelope(t, router, "/forbidden", http.StatusForbidden, 1203, "forbidden access")
	plain := assertEnvelope(t, router, "/plain", http.StatusInternalServerError, 1001, "internal server error")
	panicResponse := assertEnvelope(t, router, "/panic", http.StatusInternalServerError, 1001, "internal server error")

	if strings.Contains(string(plain.Data), "secret") || strings.Contains(string(panicResponse.Data), "secret") {
		t.Fatal("internal error details leaked in response data")
	}
	assertInternalErrorsLogged(t, logs.All())
}

func assertEnvelope(t *testing.T, router http.Handler, path string, status, code int, msg string) responseEnvelope {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != status {
		t.Fatalf("path=%s status=%d body=%s", path, rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("path=%s leaked internal details: %s", path, rec.Body.String())
	}

	var envelope responseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("path=%s decode response: %v; body=%s", path, err, rec.Body.String())
	}
	if envelope.Code != code || envelope.Msg != msg {
		t.Fatalf("path=%s envelope=%+v", path, envelope)
	}
	if envelope.RequestID == "" || rec.Header().Get("X-Request-ID") != envelope.RequestID {
		t.Fatalf("path=%s header=%q request_id=%q", path, rec.Header().Get("X-Request-ID"), envelope.RequestID)
	}
	return envelope
}

func assertJSONData(t *testing.T, raw json.RawMessage, want map[string]any) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("data=%v want=%v", got, want)
	}
}

func assertInternalErrorsLogged(t *testing.T, entries []observer.LoggedEntry) {
	t.Helper()
	var plainLogged, panicLogged, stackLogged bool
	for _, entry := range entries {
		fields := entry.ContextMap()
		errorText := fmt.Sprint(fields["error"])
		plainLogged = plainLogged || errorText == "secret"
		panicLogged = panicLogged || errorText == "secret panic"
		stackLogged = stackLogged || fmt.Sprint(fields["stack"]) != "<nil>"
	}
	if !plainLogged || !panicLogged || !stackLogged {
		t.Fatalf("plain_logged=%v panic_logged=%v stack_logged=%v entries=%v", plainLogged, panicLogged, stackLogged, entries)
	}
}
