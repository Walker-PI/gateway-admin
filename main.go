package main

import (
	"flag"
	"time"

	"github.com/Walker-PI/gateway-admin/conf"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func main() {

	var confFilePath string
	flag.StringVar(&confFilePath, "conf", "conf/app.ini", "Specify configuration file path")
	flag.Parse()

	conf.LoadConfig(confFilePath)
	logger.InitLogs()
	storage.InitStorage()

	gin.SetMode(conf.Server.RunMode)

	r := gin.New()
	r.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(zap.L(), true))

	registerRouter(r)

	_ = r.Run(":" + conf.Server.Port)
}
