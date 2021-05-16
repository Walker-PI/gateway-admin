package main

import (
	"github.com/Walker-PI/gateway-admin/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerRouter(r *gin.Engine) {

	r.GET("/ping", handler.Ping)
	// your code

}
