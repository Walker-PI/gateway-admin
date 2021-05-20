package handler

import (
	"encoding/json"
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
	TargetStripPrefix int32  `form:"target_strip_prefix" json:"target_strip_prefix"`
	TargetLb          string `form:"target_lb" json:"target_lb"`
	TargetTimeout     int64  `form:"target_timeout" json:"target_timeout"`
	MaxQPS            int32  `form:"max_qps" json:"max_qps"`
	Auth              string `form:"auth" json:"auth"`
	IPWhiteList       string `form:"ip_white_list" json:"ip_white_list"`
	IPBlackList       string `form:"ip_black_list" json:"ip_black_list"`
	Description       string `form:"column:description" json:"description"`
}

type updateAPIHandler struct {
	Ctx              *gin.Context
	Params           UpdateAPIParams
	TargetURL        *url.URL
	IPWhiteList      []net.IP
	IPBlackList      []net.IP
	APIConfig        *dal.APIGatewayConfig
	APIConfigHistory *dal.APIGatewayConfigHistory
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

	h.APIConfig, err = dal.GetAPIConfigByID(storage.MysqlClient, h.Params.APIID)
	if err != nil {
		return err
	}
	if h.APIConfig == nil {
		logger.Warn("[updateAPIHandler-Process] api_id is invalid: api_id=%v", h.Params.APIID)
		return fmt.Errorf("api_id is not exsit")
	}

	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
	db := storage.MysqlClient.Begin()
	defer func() {
		if err != nil {
			logger.Error("[updateAPIHandler-Process] update api failed: err=%v", err)
			db.Rollback()
		} else {
			db.Commit()
			logger.Info("[updatePIHandler-Process] update api succeed")
		}
	}()

	h.packAPIConfig()
	err = dal.UpdateAPI(db, h.Params.APIID, h.APIConfig)
	if err != nil {
		logger.Error("[updateAPIHandler-Process] update a api failed: err=%v", err)
		return err
	}

	h.packAPIConfigHistory()
	err = dal.CreateAPIHistory(db, h.APIConfigHistory)
	if err != nil {
		logger.Error("[updateAPIHandler-Process] update a api history failed: err=%v", err)
		return err
	}

	// APIConfig: Write to Redis
	msgBytes, err := json.Marshal(h.APIConfig)
	if err != nil {
		logger.Error("[updateAPIHandler-Process] marshal failed: err=%v", err)
		return err
	}
	key := fmt.Sprintf(constdef.APIConfigKeyFmt, h.APIConfig.ID)
	err = storage.RedisClient.Set(h.Ctx.Request.Context(), key, string(msgBytes), 3*30*24*time.Hour).Err()
	if err != nil {
		logger.Error("[updateAPIHandler-Process] write to redis failed: err=%v", err)
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

func (h *updateAPIHandler) packAPIConfig() {
	if h.Params.Pattern != "" {
		h.APIConfig.Pattern = h.Params.Pattern
	}
	if h.Params.Method != "" {
		h.APIConfig.Method = h.Params.Method
	}
	if h.Params.APIName != "" {
		h.APIConfig.APIName = h.Params.APIName
	}
	if h.Params.MaxQPS != 0 {
		h.APIConfig.MaxQPS = h.Params.MaxQPS
	}
	if h.Params.Auth != "" {
		h.APIConfig.Auth = h.Params.Auth
	}
	if h.Params.Description != "" {
		h.APIConfig.Description = h.Params.Description
	}
	if h.Params.TargetStripPrefix > 0 {
		h.APIConfig.TargetStripPrefix = h.Params.TargetStripPrefix
	}
	if h.Params.TargetMode == constdef.DefaultTargetMode {
		h.APIConfig.TargetHost = h.TargetURL.Host
		h.APIConfig.TargetScheme = h.TargetURL.Scheme
		h.APIConfig.TargetPath = h.TargetURL.Path
		h.APIConfig.TargetLb = ""
		h.APIConfig.TargetServiceName = ""
	} else if h.Params.TargetMode == constdef.ConsulTargetMode {
		h.APIConfig.TargetLb = h.Params.TargetLb
		h.APIConfig.TargetServiceName = h.Params.TargetServiceName
		h.APIConfig.TargetHost = ""
		h.APIConfig.TargetScheme = ""
		h.APIConfig.TargetPath = ""
	}
	ipBlackList := make([]string, 0)
	for _, ip := range h.IPBlackList {
		ipBlackList = append(ipBlackList, ip.To4().String())
	}
	if len(ipBlackList) > 0 {
		h.APIConfig.IPBlackList = strings.Join(ipBlackList, ",")
	}
	ipWhiteList := make([]string, 0)
	for _, ip := range h.IPWhiteList {
		ipWhiteList = append(ipWhiteList, ip.To4().String())
	}
	if len(ipWhiteList) > 0 {
		h.APIConfig.IPWhiteList = strings.Join(ipWhiteList, ",")
	}
	h.APIConfig.ModifiedTime = time.Time{}
}

func (h *updateAPIHandler) packAPIConfigHistory() {
	h.APIConfigHistory = &dal.APIGatewayConfigHistory{
		APIID:             h.APIConfig.ID,
		Pattern:           h.APIConfig.Pattern,
		Method:            h.APIConfig.Method,
		APIName:           h.APIConfig.APIName,
		TargetMode:        h.APIConfig.TargetMode,
		TargetHost:        h.APIConfig.TargetHost,
		TargetScheme:      h.APIConfig.TargetScheme,
		TargetPath:        h.APIConfig.TargetPath,
		TargetServiceName: h.APIConfig.TargetServiceName,
		TargetLb:          h.APIConfig.TargetLb,
		MaxQPS:            h.APIConfig.MaxQPS,
		Auth:              h.APIConfig.Auth,
		IPWhiteList:       h.APIConfig.IPWhiteList,
		IPBlackList:       h.APIConfig.IPBlackList,
		Description:       h.APIConfig.Description,
	}
}
