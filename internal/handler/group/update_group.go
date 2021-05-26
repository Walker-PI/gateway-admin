package group

import (
	"fmt"

	"github.com/Walker-PI/gateway-admin/internal/dal"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"github.com/gin-gonic/gin"
)

type UpdateGroupParams struct {
	GroupID     int64  `form:"group_id" json:"group_id" binding:"required"` // 分组ID
	GroupName   string `form:"group_name" json:"group_name"`                // API组名
	Active      int    `form:"active" json:"active"`                        // 状态: 1-可用 2-不可用
	Delete      int    `form:"delete" json:"delete"`                        // 1-删除 不可恢复
	Description string `form:"description" json:"description"`              // 描述
}

type updateGroupHandler struct {
	Ctx    *gin.Context
	Params UpdateGroupParams
}

func NewUpdateGroupHandler(c *gin.Context) *updateGroupHandler {
	return &updateGroupHandler{
		Ctx: c,
	}
}

// @UpdateGroup 修改路由分组
// @Description 修改路由分组接口
// @Accept application/json
// @Produce application/json
// @Param object body UpdateGroupParams true "请求参数-Body"
// @Success 200 {object} model.CommonResponse
// @Router /gateway-admin/route/update_group [post]
func UpdateGroup(c *gin.Context) (out *resp.JSONOutput) {
	h := NewUpdateGroupHandler(c)
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

func (h *updateGroupHandler) CheckParams() error {
	var err error
	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[updateGroupHandler-CheckParams] params-err: err=%v", err)
		return err
	}
	return nil
}

func (h *updateGroupHandler) Process() error {
	routeGroup, err := dal.GetRouteGroupByID(h.Params.GroupID)
	if err != nil {
		return err
	}
	if routeGroup == nil {
		err = fmt.Errorf("routeGroup isn't exsit: group_id=%v", h.Params.GroupID)
		return err
	}
	if h.Params.GroupName != "" {
		routeGroup.GroupName = h.Params.GroupName
	}
	if h.Params.Active == 1 {
		routeGroup.Status = 1
	}
	if h.Params.Active == 2 {
		routeGroup.Status = 0
	}
	if h.Params.Delete == 1 {
		routeGroup.Deleted = 1
	}
	if h.Params.Description != "" {
		routeGroup.Description = h.Params.Description
	}
	err = dal.UpdateRouteGroup(storage.MysqlClient, h.Params.GroupID, routeGroup)
	if err != nil {
		return err
	}
	return nil
}
