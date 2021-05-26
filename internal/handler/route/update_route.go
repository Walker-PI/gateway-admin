package route

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
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"github.com/Walker-PI/gateway-admin/pkg/tools"
	"github.com/gin-gonic/gin"
)

type UpdateRouteParams struct {
	RouteID              int64  `form:"route_id" json:"route_id" binding:"required"`          // 路由ID
	Methods              string `form:"methods" json:"methods"`                               // 请求方法，支持多个，以","隔开,如 GET,POST
	Pattern              string `form:"pattern" json:"pattern"`                               // Pattern
	RateLimit            int32  `form:"rate_limit" json:"rate_limit"`                         // 限流，最大QPS
	AuthType             string `form:"auth_type" json:"auth_type"`                           // 鉴权类型
	IPWhiteList          string `form:"ip_white_list" json:"ip_white_list"`                   // IP白名单
	IPBlackList          string `form:"ip_black_list" json:"ip_black_list"`                   // IP黑名单
	TargetURL            string `form:"target_url" json:"target_url"`                         // 转发URL, 非服务发现转发
	TargetTimeout        int32  `form:"target_timeout" json:"target_timeout"`                 // 目标服务超时时间, 单位ms
	Discovery            string `form:"discovery" json:"discovery"`                           // 转发模式 CONSUL EUREKA
	DiscoveryPath        string `form:"discovery_path" json:"discovery_path"`                 // 转发路径, 未配置使用原请求路径
	DiscoveryServiceName string `form:"discovery_service_name" json:"discovery_service_name"` // 目标服务名
	DiscoveryLoadBalance string `form:"discovery_load_balance" json:"discovery_load_balance"` // 负载均衡类型
}

type updateRouteHandler struct {
	Ctx         *gin.Context
	Params      UpdateRouteParams
	TargetURL   *url.URL
	IPWhiteList []net.IP
	IPBlackList []net.IP
	RouteConfig *dal.RouteConfig
}

func buildUpdateRouteHandler(c *gin.Context) *updateRouteHandler {
	return &updateRouteHandler{
		Ctx: c,
	}
}

func UpdateRoute(c *gin.Context) (out *resp.JSONOutput) {

	h := buildUpdateRouteHandler(c)

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

func (h *updateRouteHandler) CheckParams() (err error) {

	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[updateRouteHandler-checkParams] params-err: err=%v", err)
		return err
	}
	if h.Params.Pattern != "" {
		if h.Params.Pattern[0] != '/' || h.Params.Pattern != path.Clean(h.Params.Pattern) {
			err = errors.New("http: pattern is invalid")
			logger.Error("[updateRouteHandler-checkParams] params-err: pattern=%v", h.Params.Pattern)
			return err
		}
	}

	h.Params.Methods = strings.ToUpper(h.Params.Methods)
	methods := strings.Split(h.Params.Methods, ",")
	for _, method := range methods {
		if !methodMap[method] {
			err = errors.New("http: method is invalid")
			logger.Error("[updateRouteHandler-checkParams] params-err: methods=%v, error_method=%v", h.Params.Methods, method)
			return err
		}
	}

	h.Params.Discovery = strings.ToUpper(h.Params.Discovery)
	if h.Params.Discovery == "" && h.Params.TargetURL != "" {
		h.TargetURL, err = url.Parse(h.Params.TargetURL)
		if err != nil {
			logger.Error("[updateRouteHandler-checkParams] params-err: target_url=%v", h.Params.TargetURL)
			return err
		}
	} else if !discoveryMap[h.Params.Discovery] {
		err = fmt.Errorf("[updateRouteHandler-checkParams] params-err: discovery=%v", h.Params.Discovery)
		return err
	} else if discoveryMap[h.Params.Discovery] {
		if h.Params.DiscoveryServiceName == "" {
			err = errors.New("discovery: service_name is invalid")
			logger.Error("[updateRouteHandler-checkParams] params-err: service_name=%v", h.Params.DiscoveryServiceName)
			return err
		}
		h.Params.DiscoveryLoadBalance = strings.ToUpper(h.Params.DiscoveryLoadBalance)

		if h.Params.DiscoveryLoadBalance != "" && !loadBalanceMap[h.Params.DiscoveryLoadBalance] {
			err = errors.New("discovery: loadbalance is invalid")
			logger.Error("[updateRouteHandler-checkParams] params-err: lb=%v", h.Params.DiscoveryLoadBalance)
			return err
		}

		if h.Params.DiscoveryPath != "" && h.Params.DiscoveryPath != path.Clean(h.Params.DiscoveryPath) {
			err = errors.New("http: discoveryPath is invalid")
			logger.Error("[updateRouteHandler-checkParams] params-err: discoveryPath=%v", h.Params.DiscoveryPath)
			return err
		}
	}

	h.Params.AuthType = strings.ToUpper(h.Params.AuthType)

	if h.Params.AuthType != "" && !authTypeMap[h.Params.AuthType] {
		err = errors.New("auth: auth type is invalid")
		logger.Error("[createRouteHandler-checkParams] params-err: auth=%v", h.Params.AuthType)
		return err
	}

	if h.Params.IPWhiteList != "" {
		h.IPWhiteList, err = tools.GetNetIPList(h.Params.IPWhiteList)
		if err != nil {
			logger.Error("[createRouteHandler-checkParams] params-err: ip_white_list=%v", h.Params.IPWhiteList)
			return err
		}
	}
	if h.Params.IPBlackList != "" {
		h.IPBlackList, err = tools.GetNetIPList(h.Params.IPBlackList)
		if err != nil {
			logger.Error("[createRouteHandler-checkParams] params-err: ip_black_list=%v", h.Params.IPBlackList)
			return err
		}
	}
	return
}

func (h *updateRouteHandler) GetRouteConfig() (err error) {
	h.RouteConfig, err = dal.GetRouteConfigByID(h.Params.RouteID)
	if err != nil {
		logger.Error("[updateRouteHandler-GetRouteConfig] failed: err=%v", err)
		return
	}
	return nil
}

func (h *updateRouteHandler) Process() (err error) {

	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
	db := storage.MysqlClient.Begin()
	defer func() {
		if err != nil {
			logger.Error("[updateRouteHandler-Process] update routeConfig failed: err=%v", err)
			db.Rollback()
		} else {
			db.Commit()
			logger.Info("[updateRouteHandler-Process] update routeConfig succeed")
		}
	}()

	// RouteConfig: Write to Mysql
	h.packRouteConfig()
	err = dal.UpdateRouteConfig(db, h.RouteConfig.ID, h.RouteConfig)
	if err != nil {
		logger.Error("[updateRouteHandler-Process] update routeConfig failed: err=%v", err)
		return err
	}

	// 增加路由配置id到redis集合
	key := fmt.Sprintf(constdef.AllRouteConfigIDFmt, h.RouteConfig.Source)
	err = storage.RedisClient.SAdd(h.Ctx.Request.Context(), key, h.RouteConfig.ID).Err()
	if err != nil {
		logger.Error("[updateRouteHandler-Process] save route_id to redis set failed: err=%v", err)
		return err
	}

	// RouteConfig: Write to Redis
	msgBytes, err := json.Marshal(h.RouteConfig)
	if err != nil {
		logger.Error("[updateRouteHandler-Process] marshal failed: err=%v", err)
		return err
	}
	key = fmt.Sprintf(constdef.RouteConfigKeyFmt, h.RouteConfig.ID, h.RouteConfig.Source)
	err = storage.RedisClient.Set(h.Ctx.Request.Context(), key, string(msgBytes), 3*30*24*time.Hour).Err()
	if err != nil {
		logger.Error("[updateRouteHandler-Process] write to redis failed: err=%v", err)
		return err
	}
	return
}

func (h *updateRouteHandler) packRouteConfig() {
	if h.Params.Methods != "" {
		h.RouteConfig.Methods = h.Params.Methods
	}
	if h.Params.Pattern != "" {
		h.RouteConfig.Pattern = h.Params.Pattern
	}

	if h.Params.IPBlackList != "" {
		h.RouteConfig.IPBlackList = h.Params.IPBlackList
	}
	if h.Params.IPWhiteList != "" {
		h.RouteConfig.IPWhiteList = h.Params.IPWhiteList
	}
	if h.Params.RateLimit > 0 {
		h.RouteConfig.RateLimit = h.Params.RateLimit
	}

	if h.Params.TargetTimeout > 0 {
		h.RouteConfig.TargetTimeout = h.Params.TargetTimeout
	}

	if h.Params.AuthType != "" {
		h.RouteConfig.AuthType = h.Params.AuthType
	}

	if h.Params.Discovery == "" && h.Params.TargetURL != "" {
		h.RouteConfig.Discovery = h.Params.Discovery
		h.RouteConfig.TargetURL = h.Params.TargetURL
		h.RouteConfig.DiscoveryServiceName = ""
		h.RouteConfig.DiscoveryPath = ""
		h.RouteConfig.DiscoveryLoadBalance = ""
	}

	if h.Params.Discovery != "" {
		h.RouteConfig.TargetURL = ""
		h.RouteConfig.Discovery = h.Params.Discovery
		h.RouteConfig.DiscoveryPath = h.Params.DiscoveryPath
		h.RouteConfig.DiscoveryServiceName = h.Params.DiscoveryServiceName
		h.RouteConfig.DiscoveryLoadBalance = h.Params.DiscoveryLoadBalance
	}
}
