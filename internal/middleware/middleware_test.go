package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMiddlewareInterface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	tests := []struct {
		name       string
		middleware Middleware
		identifier string
	}{
		{
			name:       "request_id middleware",
			middleware: NewRequestIDMiddleware(),
			identifier: "request_id",
		},
		{
			name:       "logging middleware",
			middleware: NewLoggingMiddleware(logger),
			identifier: "logging",
		},
		{
			name:       "recovery middleware",
			middleware: NewRecoveryMiddleware(),
			identifier: "recovery",
		},
		{
			name:       "response middleware",
			middleware: NewResponseMiddleware(),
			identifier: "response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.identifier, tt.middleware.Identifier())
			assert.NotNil(t, tt.middleware.HandlerFunc())
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewRequestIDMiddleware().HandlerFunc())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// Check that response has X-Request-ID header
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	router := gin.New()
	router.Use(NewLoggingMiddleware(logger).HandlerFunc())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewRecoveryMiddleware().HandlerFunc())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestResponseMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewResponseMiddleware().HandlerFunc())
	router.GET("/success", func(c *gin.Context) {
		c.Set("data", gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestRegisterDefaultMiddlewares(t *testing.T) {
	logger := zap.NewNop()
	middlewares := RegisterDefaultMiddlewares(logger)

	assert.Len(t, middlewares, 4)
	assert.Equal(t, "request_id", middlewares[0].Identifier())
	assert.Equal(t, "logging", middlewares[1].Identifier())
	assert.Equal(t, "recovery", middlewares[2].Identifier())
	assert.Equal(t, "response", middlewares[3].Identifier())
}
