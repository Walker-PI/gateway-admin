package group

import (
	"fmt"
	"strings"

	"github.com/Walker-PI/gateway-admin/constdef"
	"github.com/Walker-PI/gateway-admin/internal/dal"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"github.com/gin-gonic/gin"
)

type CreateGroupParams struct {
	GroupName   string `form:"group_name" json:"group_name" binding:"required"` // API组名
	Source      string `form:"source" json:"source" binding:"required"`         // 来源: cloud-云端 edgex-边缘侧
	Description string `form:"description" json:"description"`                  // 描述
}

type createGroupHandler struct {
	Ctx    *gin.Context
	Params CreateGroupParams
}

var sourceMap = map[string]bool{
	constdef.SourceCloud: true,
	constdef.SourceEdgex: true,
}

func buildCreateGroupHandler(c *gin.Context) *createGroupHandler {
	return &createGroupHandler{
		Ctx: c,
	}
}

// CreateGroup 创建路由分组
// @Description 创建路由分组接口
// @Accept application/json
// @Produce application/json
// @Param object body CreateGroupParams true "请求参数-Body"
// @Success 200 {object} model.CommonResponse
// @Router /gateway-admin/route/create_group [post]
func CreateGroup(c *gin.Context) (out *resp.JSONOutput) {
	h := buildCreateGroupHandler(c)
	err := h.CheckParams()
	if err != nil {
		return resp.SampleJSON(c, resp.RespCodeParamsError, false)
	}
	err = h.Process()
	if err != nil {
		return resp.SampleJSON(c, resp.RespDatabaseError, false)
	}
	return resp.SampleJSON(c, resp.RespCodeSuccess, true)
}

func (h *createGroupHandler) CheckParams() error {
	var err error
	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[createGroupHandler-CheckParams] params-err: err=%v", err)
		return err
	}

	h.Params.Source = strings.ToUpper(h.Params.Source)

	if !sourceMap[h.Params.Source] {
		err = fmt.Errorf("source is invalid: source=%v", h.Params.Source)
		logger.Error("[createGroupHandler-CheckParams] params-err: err=%v", err)
		return err
	}
	return nil
}

func (h *createGroupHandler) Process() error {
	routeGroup := &dal.RouteGroup{
		GroupName:   h.Params.GroupName,
		Source:      h.Params.Source,
		Status:      1,
		Description: h.Params.Description,
	}

	err := dal.CreateGroup(storage.MysqlClient, routeGroup)
	if err != nil {
		logger.Error("[createGroupHandler-Process] create group failed: err=%v", err)
		return err
	}
	return nil
}
