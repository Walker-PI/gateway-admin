package handler

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/Walker-PI/gateway-admin/constdef"
	"github.com/Walker-PI/gateway-admin/internal/dal"
	"github.com/Walker-PI/gateway-admin/internal/logic"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"github.com/Walker-PI/gateway-admin/pkg/tools"
	"github.com/gin-gonic/gin"
)

type UpdateAPIParams struct {
	APIID             int64  `form:"api_id" json:"api_id" binding:"required"`
	APIName           string `form:"api_name" json:"api_name"`
	Pattern           string `form:"pattern" json:"pattern"`
	Method            string `form:"method" json:"method"`
	TargetMode        int32  `form:"target_mode" json:"target_mode"`
	TargetURL         string `form:"target_url" json:"target_url"`
	TargetServiceName string `form:"target_service_name" json:"target_service_name"`
	TargetLb          string `form:"target_lb" json:"target_lb"`
	TargetTimeout     int64  `form:"target_timeout" json:"target_timeout"`
	MaxQps            int32  `form:"max_qps" json:"max_qps"`
	Auth              string `form:"auth" json:"auth"`
	IPWhiteList       string `form:"ip_white_list" json:"ip_white_list"`
	IPBlackList       string `form:"ip_black_list" json:"ip_black_list"`
	Description       string `form:"column:description" json:"description"`
}

type updateAPIHandler struct {
	Ctx         *gin.Context
	Params      UpdateAPIParams
	TargetURL   *url.URL
	IPWhiteList []net.IP
	IPBlackList []net.IP
}

func buildUpdateAPIHandler(c *gin.Context) *updateAPIHandler {
	return &updateAPIHandler{
		Ctx: c,
	}
}

func UpdateAPI(c *gin.Context) (out *resp.JSONOutput) {

	h := buildUpdateAPIHandler(c)

	err := h.CheckParams()
	if err != nil {
		logger.Error("[UpdateAPI] params-error: err=%v", err)
		return resp.SampleJSON(c, resp.RespCodeParamsError, nil)
	}

	err = h.Process()
	if err != nil {
		logger.Error("[UpdateAPI] update api failed: err=%v", err)
		return resp.SampleJSON(c, resp.RespDatabaseError, nil)
	}

	err = h.Notify()
	if err != nil {
		logger.Error("[UpdateAPI] Notify failed: err=%v", err)
		return resp.SampleJSON(c, resp.RespCodeRedisError, nil)
	}
	return resp.SampleJSON(c, resp.RespCodeSuccess, nil)
}

func (h *updateAPIHandler) CheckParams() (err error) {
	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[updateAPIHandler-checkParams] params-err: err=%v", err)
		return err
	}
	if h.Params.Pattern != "" {
		if h.Params.Pattern[0] != '/' || h.Params.Pattern != path.Clean(h.Params.Pattern) {
			err = errors.New("http: pattern is invalid")
			logger.Error("[updateAPIHandler-checkParams] params-err: pattern=%v", h.Params.Pattern)
			return err
		}
	}
	if h.Params.Method != "" {
		if !methodMap[h.Params.Method] {
			err = errors.New("http: method is invalid")
			logger.Error("[updateAPIHandler-checkParams] params-err: method=%v", h.Params.Method)
			return err
		}
	}

	if h.Params.TargetMode != 0 {
		if !targetModeMap[h.Params.TargetMode] {
			err = errors.New("target_modo: mode is invalid")
			logger.Error("[updateAPIHandler-checkParams] params-err: mode=%v", h.Params.TargetMode)
			return err
		}
		if h.Params.TargetMode == constdef.DefaultTargetMode {
			h.TargetURL, err = url.Parse(h.Params.TargetURL)
			if err != nil {
				logger.Error("[updateAPIHandler-checkParams] params-err: mode=%v, target_url=%v", h.Params.TargetMode, h.Params.TargetURL)
				return err
			}
		} else if h.Params.TargetMode == constdef.ConsulTargetMode {
			if h.Params.TargetServiceName == "" {
				err = errors.New("consul: service_name is invalid")
				logger.Error("[updateAPIHandler-checkParams] params-err: service_name=%v", h.Params.TargetServiceName)
				return err
			}
			h.Params.TargetLb = strings.ToUpper(h.Params.TargetLb)
			if h.Params.TargetLb == "" {
				h.Params.TargetLb = constdef.RandLoadBalance
			}
			if !loadBalanceMap[h.Params.TargetLb] {
				err = errors.New("consul: loadbalance is invalid")
				logger.Error("[updateAPIHandler-checkParams] params-err: lb=%v", h.Params.TargetLb)
				return err
			}
		}
	}

	h.Params.Auth = strings.ToUpper(h.Params.Auth)
	if h.Params.Auth != "" {
		if !authMap[h.Params.Auth] {
			err = errors.New("auth: auth type is invalid")
			logger.Error("[updateAPIHandler-checkParams] params-err: auth=%v", h.Params.Auth)
			return err
		}
	}

	if h.Params.IPWhiteList != "" {
		h.IPWhiteList, err = tools.GetNetIPList(h.Params.IPWhiteList)
		if err != nil {
			logger.Error("[updateAPIHandler-checkParams] params-err: ip_white_list=%v", h.Params.IPWhiteList)
			return err
		}
	}
	if h.Params.IPBlackList != "" {
		h.IPBlackList, err = tools.GetNetIPList(h.Params.IPBlackList)
		if err != nil {
			logger.Error("[updateAPIHandler-checkParams] params-err: ip_black_list=%v", h.Params.IPBlackList)
			return err
		}
	}
	return nil
}

func (h *updateAPIHandler) Process() (err error) {

	apiConfig, err := dal.GetAPIConfigByID(storage.MysqlClient, h.Params.APIID)
	if err != nil {
		return err
	}
	if apiConfig == nil {
		logger.Warn("[updateAPIHandler-Process] api_id is invalid: api_id=%v", h.Params.APIID)
		return fmt.Errorf("api_id is not exsit")
	}
	if h.Params.Pattern != "" {
		apiConfig.Pattern = h.Params.Pattern
	}
	if h.Params.Method != "" {
		apiConfig.Method = h.Params.Method
	}
	if h.Params.APIName != "" {
		apiConfig.APIName = h.Params.APIName
	}
	if h.Params.MaxQps != 0 {
		apiConfig.MaxQps = h.Params.MaxQps
	}
	if h.Params.Auth != "" {
		apiConfig.Auth = h.Params.Auth
	}
	if h.Params.Description != "" {
		apiConfig.Description = h.Params.Description
	}
	if h.Params.TargetMode == constdef.DefaultTargetMode {
		apiConfig.TargetHost = h.TargetURL.Host
		apiConfig.TargetScheme = h.TargetURL.Scheme
		apiConfig.TargetPath = h.TargetURL.Path
		apiConfig.TargetLb = ""
		apiConfig.TargetServiceName = ""
	} else if h.Params.TargetMode == constdef.ConsulTargetMode {
		apiConfig.TargetLb = h.Params.TargetLb
		apiConfig.TargetServiceName = h.Params.TargetServiceName
		apiConfig.TargetHost = ""
		apiConfig.TargetScheme = ""
		apiConfig.TargetPath = ""
	}
	ipBlackList := make([]string, 0)
	for _, ip := range h.IPBlackList {
		ipBlackList = append(ipBlackList, ip.To4().String())
	}
	if len(ipBlackList) > 0 {
		apiConfig.IPBlackList = strings.Join(ipBlackList, ",")
	}
	ipWhiteList := make([]string, 0)
	for _, ip := range h.IPWhiteList {
		ipWhiteList = append(ipWhiteList, ip.To4().String())
	}
	if len(ipWhiteList) > 0 {
		apiConfig.IPWhiteList = strings.Join(ipWhiteList, ",")
	}
	apiConfig.ModifiedTime = time.Time{}

	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
	db := storage.MysqlClient.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	err = dal.UpdateAPI(db, h.Params.APIID, apiConfig)
	if err != nil {
		return err
	}

	apiConfigHistory := &dal.APIGatewayConfigHistory{
		APIID:             apiConfig.ID,
		Pattern:           apiConfig.Pattern,
		Method:            apiConfig.Method,
		APIName:           apiConfig.APIName,
		TargetMode:        apiConfig.TargetMode,
		TargetHost:        apiConfig.TargetHost,
		TargetScheme:      apiConfig.TargetScheme,
		TargetPath:        apiConfig.TargetPath,
		TargetServiceName: apiConfig.TargetServiceName,
		TargetLb:          apiConfig.TargetLb,
		MaxQps:            apiConfig.MaxQps,
		Auth:              apiConfig.Auth,
		IPWhiteList:       apiConfig.IPWhiteList,
		IPBlackList:       apiConfig.IPBlackList,
		Description:       apiConfig.Description,
	}

	err = dal.CreateAPIHistory(db, apiConfigHistory)
	if err != nil {
		return err
	}
	return nil
}

func (h *updateAPIHandler) Notify() (err error) {
	// Notify update
	for i := 0; i < 3; i++ {
		err = logic.RedisPub(h.Ctx.Request.Context(), constdef.UpdateGatewayRoute, "by update api")
		if err == nil {
			return
		}
		logger.Error("[updateAPIHandler-Notify] Notify failed: try=%v, channl=%v, message=%v",
			i+1, constdef.UpdateGatewayRoute, "by update api")
	}
	return
}
