package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

func TestMiddlewareInterface(t *testing.T) {
	Convey("Middleware 接口实现遍历", t, func() {
		gin.SetMode(gin.TestMode)
		logger := zap.NewNop()

		cases := []struct {
			name       string
			middleware Middleware
			identifier string
		}{
			{"request_id middleware", NewRequestIDMiddleware(), "request_id"},
			{"logging middleware", NewLoggingMiddleware(logger), "logging"},
			{"recovery middleware", NewRecoveryMiddleware(), "recovery"},
			{"response middleware", NewResponseMiddleware(), "response"},
		}

		for _, c := range cases {
			So(c.middleware.Identifier(), ShouldEqual, c.identifier)
			So(c.middleware.HandlerFunc(), ShouldNotBeNil)
		}
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	Convey("RequestIDMiddleware 在响应中带 X-Request-ID", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(NewRequestIDMiddleware().HandlerFunc())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 200)
		So(w.Header().Get("X-Request-ID"), ShouldNotBeEmpty)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	Convey("LoggingMiddleware 正常透传 200", t, func() {
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

		So(w.Code, ShouldEqual, 200)
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	Convey("RecoveryMiddleware 捕获 panic 返回 500", t, func() {
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.Use(NewRecoveryMiddleware().HandlerFunc())
		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 500)
	})
}

func TestResponseMiddleware(t *testing.T) {
	Convey("ResponseMiddleware 包装响应输出", t, func() {
		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.Use(NewResponseMiddleware().HandlerFunc())
		router.GET("/success", func(c *gin.Context) {
			c.Set("data", gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/success", nil)
		router.ServeHTTP(w, req)

		So(w.Code, ShouldEqual, 200)
		So(w.Body.String(), ShouldContainSubstring, "success")
	})
}

func TestRegisterDefaultMiddlewares(t *testing.T) {
	Convey("RegisterDefaultMiddlewares 返回 4 个默认中间件并顺序固定", t, func() {
		logger := zap.NewNop()
		middlewares := RegisterDefaultMiddlewares(logger)

		So(len(middlewares), ShouldEqual, 4)
		So(middlewares[0].Identifier(), ShouldEqual, "request_id")
		So(middlewares[1].Identifier(), ShouldEqual, "logging")
		So(middlewares[2].Identifier(), ShouldEqual, "recovery")
		So(middlewares[3].Identifier(), ShouldEqual, "response")
	})
}
