package main

import (
	"github.com/Walker-PI/gateway-admin/internal/handler"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/gin-gonic/gin"
)

func registerRouter(r *gin.Engine) {

	r.GET("/ping", Ping)
	// your code

	r.POST("/gateway-admin/create_api", resp.JSONOutPutWrapper(handler.CreateAPI))
	r.POST("/gateway-admin/update_api", resp.JSONOutPutWrapper(handler.UpdateAPI))
	r.GET("/gateway-admin/delete_api", resp.JSONOutPutWrapper(handler.DeleteAPI))
	r.GET("/gateway-admin/get_api")
}

func Ping(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")
	_, _ = c.Writer.Write([]byte("pong"))
}
