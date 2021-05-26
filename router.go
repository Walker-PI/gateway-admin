package main

import (
	"github.com/Walker-PI/gateway-admin/internal/handler"
	"github.com/Walker-PI/gateway-admin/internal/handler/group"
	"github.com/Walker-PI/gateway-admin/internal/handler/route"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/gin-gonic/gin"
	ginswagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "github.com/Walker-PI/gateway-admin/docs"
)

func registerRouter(r *gin.Engine) {

	r.GET("/ping", Ping)
	// your code

	// apiRouter := r.Group("/gateway-admin/api")
	// {
	// 	apiRouter.POST("/create", resp.JSONOutPutWrapper(handler.CreateAPI))
	// 	apiRouter.POST("/update", resp.JSONOutPutWrapper(handler.UpdateAPI))
	// 	apiRouter.GET("/delete", resp.JSONOutPutWrapper(handler.DeleteAPI))
	// 	apiRouter.GET("/search", resp.JSONOutPutWrapper(handler.CreateAPI))
	// }

	gatewayRouter := r.Group("/gateway-admin/route")
	{
		gatewayRouter.POST("/create_group", resp.JSONOutPutWrapper(group.CreateGroup))
		gatewayRouter.POST("/update_group", resp.JSONOutPutWrapper(group.UpdateGroup))

		gatewayRouter.POST("/create_route", resp.JSONOutPutWrapper(route.CreateRoute))
		gatewayRouter.POST("/update_route", resp.JSONOutPutWrapper(route.UpdateRoute))
		gatewayRouter.POST("/delete_route", resp.JSONOutPutWrapper(route.DeleteRoute))

		gatewayRouter.GET("/search_group", resp.JSONOutPutWrapper(group.SearchGroup)) // 查询所有的group
		gatewayRouter.GET("/search_route", resp.JSONOutPutWrapper(route.SearchRoute)) // 通过group_id 或者 group_name查所有的route_info

	}

	r.GET("/gateway-admin/internal/get_token", resp.JSONOutPutWrapper(handler.GetToken))

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerFiles.Handler))
}

func Ping(c *gin.Context) {
	c.Writer.Header().Add("Content-Type", "text/plain")
	_, _ = c.Writer.Write([]byte("pong"))
}
