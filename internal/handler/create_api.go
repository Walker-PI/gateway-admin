package handler

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Walker-PI/gateway-admin/constdef"
	"github.com/Walker-PI/gateway-admin/internal/dal"
	"github.com/Walker-PI/gateway-admin/internal/logic"
	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"github.com/Walker-PI/gateway-admin/pkg/tools"
	"github.com/gin-gonic/gin"
)

type CreateAPIParams struct {
	Pattern           string `form:"pattern" json:"pattern" binding:"required"`
	Method            string `form:"method" json:"method" binding:"required"`
	APIName           string `form:"api_name" json:"api_name" binding:"required"`
	TargetMode        int32  `form:"target_mode" json:"target_mode" binding:"required"`
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

var targetModeMap = map[int32]bool{
	constdef.DefaultTargetMode: true,
	constdef.ConsulTargetMode:  true,
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

var authMap = map[string]bool{
	constdef.KeyLess: true,
	constdef.AuthJWT: true,
}

type createAPIHandler struct {
	Ctx         *gin.Context
	Params      CreateAPIParams
	TargetURL   *url.URL
	IPWhiteList []net.IP
	IPBlackList []net.IP
}

func buildCreateAPIHandler(c *gin.Context) *createAPIHandler {
	return &createAPIHandler{
		Ctx: c,
	}
}

func CreateAPI(c *gin.Context) (out *resp.JSONOutput) {

	h := buildCreateAPIHandler(c)

	err := h.CheckParams()
	if err != nil {
		logger.Error("[CreateAPI] params-error: err=%v", err)
		return resp.SampleJSON(c, resp.RespCodeParamsError, nil)
	}

	err = h.Process()
	if err != nil {
		logger.Error("[CreateAPI] Write API to DB failed: err=%v", err)
		return resp.SampleJSON(c, resp.RespDatabaseError, nil)
	}

	err = h.Notify()
	if err != nil {
		logger.Error("[CreateAPI] Notify failed: err=%v", err)
		return resp.SampleJSON(c, resp.RespCodeRedisError, nil)
	}
	return resp.SampleJSON(c, resp.RespCodeSuccess, nil)
}

func (h *createAPIHandler) CheckParams() (err error) {
	err = h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[createAPIHandler-checkParams] params-err: err=%v", err)
		return err
	}
	if h.Params.Pattern[0] != '/' || h.Params.Pattern != path.Clean(h.Params.Pattern) {
		err = errors.New("http: pattern is invalid")
		logger.Error("[createAPIHandler-checkParams] params-err: pattern=%v", h.Params.Pattern)
		return err
	}

	if !methodMap[h.Params.Method] {
		err = errors.New("http: method is invalid")
		logger.Error("[createAPIHandler-checkParams] params-err: method=%v", h.Params.Method)
		return err
	}

	if !targetModeMap[h.Params.TargetMode] {
		err = errors.New("target_modo: mode is invalid")
		logger.Error("[createAPIHandler-checkParams] params-err: mode=%v", h.Params.TargetMode)
		return err
	}

	if h.Params.TargetMode == constdef.DefaultTargetMode {
		h.TargetURL, err = url.Parse(h.Params.TargetURL)
		if err != nil {
			logger.Error("[createAPIHandler-checkParams] params-err: mode=%v, target_url=%v", h.Params.TargetMode, h.Params.TargetURL)
			return err
		}
	} else if h.Params.TargetMode == constdef.ConsulTargetMode {
		if h.Params.TargetServiceName == "" {
			err = errors.New("consul: service_name is invalid")
			logger.Error("[createAPIHandler-checkParams] params-err: service_name=%v", h.Params.TargetServiceName)
			return err
		}
		h.Params.TargetLb = strings.ToUpper(h.Params.TargetLb)
		if h.Params.TargetLb == "" {
			h.Params.TargetLb = constdef.RandLoadBalance
		}
		if !loadBalanceMap[h.Params.TargetLb] {
			err = errors.New("consul: loadbalance is invalid")
			logger.Error("[createAPIHandler-checkParams] params-err: lb=%v", h.Params.TargetLb)
			return err
		}
	}

	h.Params.Auth = strings.ToUpper(h.Params.Auth)
	if h.Params.Auth == "" {
		h.Params.Auth = constdef.KeyLess
	}
	if !authMap[h.Params.Auth] {
		err = errors.New("auth: auth type is invalid")
		logger.Error("[createAPIHandler-checkParams] params-err: auth=%v", h.Params.Auth)
		return err
	}

	if h.Params.IPWhiteList != "" {
		h.IPWhiteList, err = tools.GetNetIPList(h.Params.IPWhiteList)
		if err != nil {
			logger.Error("[createAPIHandler-checkParams] params-err: ip_white_list=%v", h.Params.IPWhiteList)
			return err
		}
	}
	if h.Params.IPBlackList != "" {
		h.IPBlackList, err = tools.GetNetIPList(h.Params.IPBlackList)
		if err != nil {
			logger.Error("[createAPIHandler-checkParams] params-err: ip_black_list=%v", h.Params.IPBlackList)
			return err
		}
	}
	return nil
}

func (h *createAPIHandler) Process() (err error) {
	apiConfig := &dal.APIGatewayConfig{
		Pattern:       h.Params.Pattern,
		Method:        h.Params.Method,
		APIName:       h.Params.APIName,
		TargetMode:    h.Params.TargetMode,
		TargetTimeout: h.Params.TargetTimeout,
		MaxQps:        h.Params.MaxQps,
		Auth:          h.Params.Auth,
		// CreatedTime:   time.Now(),
		// ModifiedTime:  time.Now(),
		Status:      1,
		Description: h.Params.Description,
	}
	if h.Params.TargetMode == constdef.DefaultTargetMode {
		apiConfig.TargetHost = h.TargetURL.Host
		apiConfig.TargetScheme = h.TargetURL.Scheme
		apiConfig.TargetPath = h.TargetURL.Path
	} else if h.Params.TargetMode == constdef.ConsulTargetMode {
		apiConfig.TargetLb = h.Params.TargetLb
		apiConfig.TargetServiceName = h.Params.TargetServiceName
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

	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
	db := storage.MysqlClient.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	// Create API
	err = dal.CreateAPI(db, apiConfig)
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

	// Create record
	err = dal.CreateAPIHistory(db, apiConfigHistory)
	if err != nil {
		return
	}
	return
}

func (h *createAPIHandler) Notify() (err error) {
	// Notify update
	for i := 0; i < 3; i++ {
		err = logic.RedisPub(h.Ctx.Request.Context(), constdef.UpdateGatewayRoute, "by create api")
		if err == nil {
			return
		}
		logger.Error("[createAPIHandler-Notify] Notify failed: try=%v, channl=%v, message=%v",
			i+1, constdef.UpdateGatewayRoute, "by create api")
	}
	return
}
