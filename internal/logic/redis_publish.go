package logic

import (
	"context"

	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
)

// RedisPub ..。
func RedisPub(ctx context.Context, channel string, message interface{}) (err error) {
	err = storage.RedisClient.Publish(ctx, channel, message).Err()
	if err != nil {
		logger.Error("[RedisPub] publish failed: channel=%v, message=%+v", channel, message)
		return
	}
	return nil
}
