package handler

import (
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")
	_, _ = c.Writer.Write([]byte("pong"))
}
