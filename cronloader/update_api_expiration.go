package cronloader

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Walker-PI/gateway-admin/constdef"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
)

func updateAPIConfigExpiration() {

	var (
		err        error
		ctx        = context.Background()
		apiStrList []string
	)

	apiStrList, err = storage.RedisClient.SMembers(ctx, constdef.AllAPIConfigID).Result()
	if err != nil {
		logger.Error("[updateAPIConfigExpiration] get all api_id failed: err=%v", err)
		return
	}

	redisPipeline := storage.RedisClient.Pipeline()
	defer redisPipeline.Close()

	for _, apiStr := range apiStrList {
		apiID, innErr := strconv.ParseInt(apiStr, 10, 64)
		if innErr != nil {
			continue
		}
		key := fmt.Sprintf(constdef.APIConfigKeyFmt, apiID)
		redisPipeline.Expire(ctx, key, 3*30*24*time.Hour)
	}
	_, err = redisPipeline.Exec(ctx)
	if err != nil {
		logger.Error("[updateAPIConfigExpiration] redis pipeline exec failed: err=%v", err)
		return
	}
	logger.Info("[updateAPIConfigExpiration] succeed")
}
