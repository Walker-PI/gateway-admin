package main

import (
	"github.com/Walker-PI/gateway-admin/internal/handler"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/gin-gonic/gin"
)

func registerRouter(r *gin.Engine) {

	r.GET("/ping", Ping)
	// your code

	apiRouter := r.Group("/gateway/api")
	{
		apiRouter.POST("/create", resp.JSONOutPutWrapper(handler.CreateAPI))
		apiRouter.POST("/update", resp.JSONOutPutWrapper(handler.UpdateAPI))
		apiRouter.GET("/delete", resp.JSONOutPutWrapper(handler.DeleteAPI))
		apiRouter.GET("/search", resp.JSONOutPutWrapper(handler.CreateAPI))
	}

	r.GET("/gateway-admin/internal/get_token", resp.JSONOutPutWrapper(handler.GetToken))
}

func Ping(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")
	_, _ = c.Writer.Write([]byte("pong"))
}
