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

func updateRouteConfigExpiration() {

	for _, source := range []string{"CLOUD", "EDGEX"} {
		var (
			err          error
			ctx          = context.Background()
			routeStrList []string
		)
		key := fmt.Sprintf(constdef.AllRouteConfigIDFmt, source)
		routeStrList, err = storage.RedisClient.SMembers(ctx, key).Result()
		if err != nil {
			logger.Error("[updateRouteConfigExpiration] get all route_id failed: err=%v", err)
			continue
		}

		redisPipeline := storage.RedisClient.Pipeline()
		defer redisPipeline.Close()

		for _, routeStr := range routeStrList {
			routeID, innErr := strconv.ParseInt(routeStr, 10, 64)
			if innErr != nil {
				continue
			}
			key := fmt.Sprintf(constdef.RouteConfigKeyFmt, routeID, source)
			redisPipeline.Expire(ctx, key, 3*30*24*time.Hour)
		}
		_, err = redisPipeline.Exec(ctx)
		if err != nil {
			logger.Error("[updateRouteConfigExpiration] redis pipeline exec failed: err=%v", err)
			continue
		}
		logger.Info("[updateRouteConfigExpiration] succeed: source=%v", source)
	}
}
