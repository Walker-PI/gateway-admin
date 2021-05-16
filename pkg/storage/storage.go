package storage

import (
	"context"
	"fmt"

	"github.com/Walker-PI/gateway-admin/conf"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	RedisClient *redis.Client // RedisClient
	MysqlClient *gorm.DB      // MysqlClient
)

func InitStorage() {
	initRedisClient()
	initMysqlClient()
}

func initRedisClient() {
	redisOpt := &redis.Options{
		Addr:     conf.RedisConf.Address,
		Password: conf.RedisConf.Password,
		DB:       conf.RedisConf.DB,
	}
	RedisClient = redis.NewClient(redisOpt)

	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func initMysqlClient() {

	optional := conf.GetDefaultDBOptional()

	format := "%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s"
	dbConfig := fmt.Sprintf(format, optional.User, optional.Password, optional.DBHostname, optional.DBPort,
		optional.DBName, optional.DBCharset, optional.Timeout, optional.ReadTimeout, optional.WriteTimeout)

	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	var err error
	MysqlClient, err = gorm.Open(mysql.New(mysql.Config{
		DriverName: optional.DriverName,
		DSN:        dbConfig,
	}), &gormConfig)

	if err != nil {
		panic(err)
	}
}
