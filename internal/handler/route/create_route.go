package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
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

type CreateRouteParams struct {
	GroupID              int64  `form:"group_id" json:"group_id" binding:"required"`          // 路由所在组
	Methods              string `form:"methods" json:"methods" binding:"required"`            // 请求方法，支持多个，以","隔开,如 GET,POST
	Pattern              string `form:"pattern" json:"pattern" binding:"required"`            // Pattern
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

var methodMap = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodConnect: true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

var loadBalanceMap = map[string]bool{
	constdef.RandLoadBalance:       true,
	constdef.RoundRobinLoadBalance: true,
	constdef.IPHashLoadBalance:     true,
	constdef.URLHashLoadBalance:    true,
}

var authTypeMap = map[string]bool{
	constdef.KeyLess: true,
	constdef.AuthJWT: true,
}

var discoveryMap = map[string]bool{
	constdef.DiscoveryEureka: true,
	constdef.DiscoveryConsul: true,
}

type createRouteHandler struct {
	Ctx         *gin.Context
	Params      CreateRouteParams
	TargetURL   *url.URL
	IPWhiteList []net.IP
	IPBlackList []net.IP
	RouteConfig *dal.RouteConfig
	RouteGroup  *dal.RouteGroup
}

func buildCreateRouteHandler(c *gin.Context) *createRouteHandler {
	return &createRouteHandler{
		Ctx: c,
	}
}

// @CreateRoute 创建API
// @Description 创建API的接口
// @Accept application/json
// @Produce application/json
// @Param object body CreateRouteParams true "请求body"
// @Success 200 {object} model.CommonResponse
// @Router /gateway-admin/api/create [post]
func CreateRoute(c *gin.Context) (out *resp.JSONOutput) {

	h := buildCreateRouteHandler(c)

	err := h.CheckParams()
	if err != nil {
		return resp.SampleJSON(c, resp.RespCodeParamsError, false)
	}

	err = h.GetRouteGroup()
	if err != nil {
		return resp.SampleJSON(c, resp.RespDatabaseError, false)
	}

	err = h.Process()
	if err != nil {
		return resp.SampleJSON(c, resp.RespDatabaseError, false)
	}

	err = h.Notify()
	if err != nil {
		return resp.SampleJSON(c, resp.RespCodeRedisError, false)
	}
	return resp.SampleJSON(c, resp.RespCodeSuccess, true)
}

func (h *createRouteHandler) CheckParams() (err error) {
	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[createRouteHandler-checkParams] params-err: err=%v", err)
		return err
	}
	if h.Params.Pattern[0] != '/' || h.Params.Pattern != path.Clean(h.Params.Pattern) {
		err = errors.New("http: pattern is invalid")
		logger.Error("[createRouteHandler-checkParams] params-err: pattern=%v", h.Params.Pattern)
		return err
	}

	h.Params.Methods = strings.ToUpper(h.Params.Methods)
	methods := strings.Split(h.Params.Methods, ",")
	for _, method := range methods {
		if !methodMap[method] {
			err = errors.New("http: method is invalid")
			logger.Error("[createRouteHandler-checkParams] params-err: methods=%v, error_method=%v", h.Params.Methods, method)
			return err
		}
	}
	h.Params.Discovery = strings.ToUpper(h.Params.Discovery)
	if h.Params.Discovery == "" {
		h.TargetURL, err = url.Parse(h.Params.TargetURL)
		if err != nil {
			logger.Error("[createRouteHandler-checkParams] params-err: target_url=%v", h.Params.TargetURL)
			return err
		}
	} else if !discoveryMap[h.Params.Discovery] {
		err = fmt.Errorf("[createRouteHandler-checkParams] params-err: discovery=%v", h.Params.Discovery)
		return err
	} else if discoveryMap[h.Params.Discovery] {
		if h.Params.DiscoveryServiceName == "" {
			err = errors.New("discovery: service_name is invalid")
			logger.Error("[createRouteHandler-checkParams] params-err: service_name=%v", h.Params.DiscoveryServiceName)
			return err
		}
		h.Params.DiscoveryLoadBalance = strings.ToUpper(h.Params.DiscoveryLoadBalance)
		if h.Params.DiscoveryLoadBalance == "" {
			h.Params.DiscoveryLoadBalance = constdef.RandLoadBalance
		}
		if !loadBalanceMap[h.Params.DiscoveryLoadBalance] {
			err = errors.New("discovery: loadbalance is invalid")
			logger.Error("[createRouteHandler-checkParams] params-err: lb=%v", h.Params.DiscoveryLoadBalance)
			return err
		}

		if h.Params.DiscoveryPath != "" && h.Params.DiscoveryPath != path.Clean(h.Params.DiscoveryPath) {
			err = errors.New("http: discoveryPath is invalid")
			logger.Error("[createRouteHandler-checkParams] params-err: discoveryPath=%v", h.Params.DiscoveryPath)
			return err
		}
	}

	h.Params.AuthType = strings.ToUpper(h.Params.AuthType)
	if h.Params.AuthType == "" {
		h.Params.AuthType = constdef.KeyLess
	}

	if !authTypeMap[h.Params.AuthType] {
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
	return nil
}

func (h *createRouteHandler) GetRouteGroup() (err error) {
	h.RouteGroup, err = dal.GetRouteGroupByID(h.Params.GroupID)
	if err != nil {
		return
	}
	if h.RouteGroup == nil {
		err = fmt.Errorf("group_id is invalid:  group_id=%v", h.Params.GroupID)
		return
	}
	return err
}

func (h *createRouteHandler) Process() (err error) {

	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
	db := storage.MysqlClient.Begin()
	defer func() {
		if err != nil {
			logger.Error("[createRouteHandler-Process] create routeConfig failed: err=%v", err)
			db.Rollback()
		} else {
			db.Commit()
			logger.Info("[createRouteHandler-Process] create routeConfig succeed")
		}
	}()

	// RouteConfig: Write to Mysql
	h.packRouteConfig()
	err = dal.CreateRouteConfig(db, h.RouteConfig)
	if err != nil {
		logger.Error("[createRouteHandler-Process] write routeConfig failed: err=%v", err)
		return err
	}

	// 增加路由配置id到redis集合
	key := fmt.Sprintf(constdef.AllRouteConfigIDFmt, h.RouteGroup.Source)
	err = storage.RedisClient.SAdd(h.Ctx.Request.Context(), key, h.RouteConfig.ID).Err()
	if err != nil {
		logger.Error("[createRouteHandler-Process] save route_id to redis set failed: err=%v", err)
		return err
	}

	// RouteConfig: Write to Redis
	msgBytes, err := json.Marshal(h.RouteConfig)
	if err != nil {
		logger.Error("[createRouteHandler-Process] marshal failed: err=%v", err)
		return err
	}
	key = fmt.Sprintf(constdef.RouteConfigKeyFmt, h.RouteConfig.ID, h.RouteConfig.Source)
	err = storage.RedisClient.Set(h.Ctx.Request.Context(), key, string(msgBytes), 3*30*24*time.Hour).Err()
	if err != nil {
		logger.Error("[createRouteHandler-Process] write to redis failed: err=%v", err)
		return err
	}
	return
}

func (h *createRouteHandler) Notify() (err error) {
	channel := h.RouteConfig.Source + "-" + constdef.UpdateGatewayRoute
	// Notify update
	for i := 0; i < 3; i++ {
		err = logic.RedisPub(h.Ctx.Request.Context(), channel, "by create api")
		if err == nil {
			return
		}
		logger.Error("[createRouteHandler-Notify] Notify failed: try=%v, channl=%v, message=%v",
			i+1, channel, "by create api")
	}
	return
}

func (h *createRouteHandler) packRouteConfig() {
	h.RouteConfig = &dal.RouteConfig{
		GroupID:       h.RouteGroup.ID,
		GroupName:     h.RouteGroup.GroupName,
		Source:        h.RouteGroup.Source,
		Pattern:       h.Params.Pattern,
		Methods:       h.Params.Methods,
		RateLimit:     h.Params.RateLimit,
		AuthType:      h.Params.AuthType,
		Discovery:     h.Params.Discovery,
		TargetTimeout: h.Params.TargetTimeout,
	}
	if h.Params.Discovery == "" {
		h.RouteConfig.TargetURL = h.TargetURL.String()
	} else {
		h.RouteConfig.DiscoveryPath = h.Params.DiscoveryPath
		h.RouteConfig.DiscoveryServiceName = h.Params.DiscoveryServiceName
		h.RouteConfig.DiscoveryLoadBalance = h.Params.DiscoveryLoadBalance
	}
	ipBlackList := make([]string, 0)
	for _, ip := range h.IPBlackList {
		ipBlackList = append(ipBlackList, ip.To4().String())
	}
	if len(ipBlackList) > 0 {
		h.RouteConfig.IPBlackList = strings.Join(ipBlackList, ",")
	}
	ipWhiteList := make([]string, 0)
	for _, ip := range h.IPWhiteList {
		ipWhiteList = append(ipWhiteList, ip.To4().String())
	}
	if len(ipWhiteList) > 0 {
		h.RouteConfig.IPWhiteList = strings.Join(ipWhiteList, ",")
	}
}
