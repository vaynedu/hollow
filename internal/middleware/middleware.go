package middleware

import "github.com/gin-gonic/gin"

type Middleware interface {
	HandlerFunc() gin.HandlerFunc
	Identifier() string
}
