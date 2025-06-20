package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/example/service"
)

func HelloHandler(c *gin.Context) {
	res := service.Hello("hello")
	c.Set("data", gin.H{"message": res})
}
