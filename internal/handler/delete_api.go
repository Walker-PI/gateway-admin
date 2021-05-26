package handler

// import (
// 	"github.com/Walker-PI/gateway-admin/constdef"
// 	"github.com/Walker-PI/gateway-admin/internal/dal"
// 	"github.com/Walker-PI/gateway-admin/internal/logic"
// 	"github.com/Walker-PI/gateway-admin/pkg/logger"
// 	"github.com/Walker-PI/gateway-admin/pkg/resp"
// 	"github.com/Walker-PI/gateway-admin/pkg/storage"
// 	"github.com/gin-gonic/gin"
// )

// type DeleteAPIParams struct {
// 	APIID int64 `form:"api_id" json:"api_id" binding:"api_id"`
// }

// type deleteAPIHandler struct {
// 	Ctx    *gin.Context
// 	Params DeleteAPIParams
// }

// func buildDeleteAPIHandler(c *gin.Context) *deleteAPIHandler {
// 	return &deleteAPIHandler{
// 		Ctx: c,
// 	}
// }

// func DeleteAPI(c *gin.Context) (out *resp.JSONOutput) {

// 	h := buildDeleteAPIHandler(c)

// 	err := h.CheckParams()
// 	if err != nil {
// 		logger.Error("[DeleteAPI] CheckParams failed: err=%v", err)
// 		return resp.SampleJSON(c, resp.RespCodeParamsError, nil)
// 	}

// 	err = h.Process()
// 	if err != nil {
// 		logger.Error("[DeleteAPI] delete api failed: err=%v", err)
// 		return resp.SampleJSON(c, resp.RespDatabaseError, nil)
// 	}

// 	err = h.Notify()
// 	if err != nil {
// 		logger.Error("[DeleteAPI] Notify failed: err=%v", err)
// 		return resp.SampleJSON(c, resp.RespCodeRedisError, nil)
// 	}

// 	return resp.SampleJSON(c, resp.RespCodeSuccess, nil)
// }

// func (h *deleteAPIHandler) CheckParams() (err error) {
// 	err = h.Ctx.Bind(&h.Params)
// 	if err != nil {
// 		logger.Error("[deleteAPIHandler-CheckParams] param-err: err=%v", err)
// 		return
// 	}
// 	return nil
// }

// func (h *deleteAPIHandler) Process() (err error) {

// 	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
// 	db := storage.MysqlClient.Begin()
// 	defer func() {
// 		if err != nil {
// 			logger.Error("[deleteAPIHandler-Process] delete api failed: err=%v", err)
// 			db.Rollback()
// 		} else {
// 			db.Commit()
// 			logger.Info("[deletePIHandler-Process] delete api succeed")
// 		}
// 	}()

// 	err = dal.DeleteAPI(db, h.Params.APIID)
// 	if err != nil {
// 		logger.Error("[deleteAPIHandler] delete api failed: err=%v", err)
// 		return err
// 	}

// 	err = storage.RedisClient.SRem(h.Ctx.Request.Context(), constdef.AllAPIConfigID, h.Params.APIID).Err()
// 	if err != nil {
// 		logger.Error("[deleteAPIHandler-Process] delete api_id from redis set failed: err=%v", err)
// 		return err
// 	}

// 	return
// }

// func (h *deleteAPIHandler) Notify() (err error) {
// 	// Notify update
// 	for i := 0; i < 3; i++ {
// 		err = logic.RedisPub(h.Ctx.Request.Context(), constdef.UpdateGatewayRoute, "by delete api")
// 		if err == nil {
// 			return
// 		}
// 		logger.Error("[deleteAPIHandler-Notify] Notify failed: try=%v, channl=%v, message=%v",
// 			i+1, constdef.UpdateGatewayRoute, "by delete api")
// 	}
// 	return
// }
