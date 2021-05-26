package route

import (
	"fmt"

	"github.com/Walker-PI/gateway-admin/constdef"
	"github.com/Walker-PI/gateway-admin/internal/dal"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"github.com/gin-gonic/gin"
)

type DeleteRouteParams struct {
	RouteID int64 `form:"route_id" json:"route_id" binding:"required"`
}

type deleteRouteHandler struct {
	Ctx         *gin.Context
	Params      DeleteRouteParams
	RouteConfig *dal.RouteConfig
}

func buildDeleteRouteHandler(c *gin.Context) *deleteRouteHandler {
	return &deleteRouteHandler{
		Ctx: c,
	}
}

func DeleteRoute(c *gin.Context) (out *resp.JSONOutput) {

	h := buildDeleteRouteHandler(c)

	err := h.CheckParams()
	if err != nil {
		return resp.SampleJSON(c, resp.RespCodeParamsError, false)
	}

	err = h.GetRouteConfig()
	if err != nil {
		return resp.SampleJSON(c, resp.RespDatabaseError, false)
	}

	err = h.Process()
	if err != nil {
		return resp.SampleJSON(c, resp.RespDatabaseError, false)
	}

	return resp.SampleJSON(c, resp.RespCodeSuccess, true)
}

func (h *deleteRouteHandler) CheckParams() (err error) {
	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[deleteRouteHandler-CheckParams] params-err: err=%v", err)
		return
	}
	return
}

func (h *deleteRouteHandler) GetRouteConfig() (err error) {
	h.RouteConfig, err = dal.GetRouteConfigByID(h.Params.RouteID)
	if err != nil {
		logger.Error("[deleteRouteHandler-GetRouteConfig] failed: err=%v", err)
		return
	}
	return nil
}

func (h *deleteRouteHandler) Process() (err error) {

	// 开启数据库事务
	db := storage.MysqlClient.Begin()
	defer func() {
		if err != nil {
			logger.Error("[deleteRouteHandler-Process] delete routeConfig failed: err=%v", err)
			db.Rollback()
		} else {
			db.Commit()
			logger.Info("[deleteRouteHandler-Process] delete routeConfig succeed")
		}
	}()

	// RouteConfig: Delete
	err = dal.DeleteRouteConfig(db, h.RouteConfig.ID)
	if err != nil {
		logger.Error("[deleteRouteHandler-Process] delete routeConfig failed: err=%v", err)
		return err
	}

	// Redis中删除路由id
	key := fmt.Sprintf(constdef.AllRouteConfigIDFmt, h.RouteConfig.Source)
	err = storage.RedisClient.SRem(h.Ctx.Request.Context(), key, h.RouteConfig.ID).Err()
	if err != nil {
		logger.Error("[deleteRouteHandler-Process] remove route_id from redis set failed: err=%v", err)
		return err
	}
	key = fmt.Sprintf(constdef.RouteConfigKeyFmt, h.RouteConfig.ID, h.RouteConfig.Source)
	err = storage.RedisClient.Del(h.Ctx.Request.Context(), key).Err()
	if err != nil {
		logger.Error("[deleteRouteHandler-Process] del from redis failed: err=%v", err)
		return err
	}
	return nil
}
