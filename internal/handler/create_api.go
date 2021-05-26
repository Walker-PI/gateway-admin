package handler

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net"
// 	"net/http"
// 	"net/url"
// 	"path"
// 	"strings"
// 	"time"

// 	"github.com/Walker-PI/gateway-admin/constdef"
// 	"github.com/Walker-PI/gateway-admin/internal/dal"
// 	"github.com/Walker-PI/gateway-admin/internal/logic"
// 	"github.com/Walker-PI/gateway-admin/pkg/logger"
// 	"github.com/Walker-PI/gateway-admin/pkg/resp"
// 	"github.com/Walker-PI/gateway-admin/pkg/storage"
// 	"github.com/Walker-PI/gateway-admin/pkg/tools"
// 	"github.com/gin-gonic/gin"
// )

// type CreateAPIParams struct {
// 	Pattern           string `form:"pattern" json:"pattern" binding:"required"`         // Pattern
// 	Method            string `form:"method" json:"method" binding:"required"`           // Method
// 	APIName           string `form:"api_name" json:"api_name" binding:"required"`       // API名字
// 	TargetMode        int32  `form:"target_mode" json:"target_mode" binding:"required"` // API转发模式
// 	TargetURL         string `form:"target_url" json:"target_url"`                      // 目标URL
// 	TargetServiceName string `form:"target_service_name" json:"target_service_name"`    // 目标服务名
// 	TargetStripPrefix int32  `form:"target_strip_prefix" json:"target_strip_prefix"`    // 是否去掉Pattern前缀
// 	TargetLb          string `form:"target_lb" json:"target_lb"`                        // 负载均衡类型
// 	TargetTimeout     int64  `form:"target_timeout" json:"target_timeout"`              // 目标服务超市时间
// 	MaxQPS            int32  `form:"max_qps" json:"max_qps"`                            // 限流，最大QPS
// 	Auth              string `form:"auth" json:"auth"`                                  // 鉴权类型
// 	IPWhiteList       string `form:"ip_white_list" json:"ip_white_list"`                // IP白名单
// 	IPBlackList       string `form:"ip_black_list" json:"ip_black_list"`                // IP黑名单
// 	Description       string `form:"column:description" json:"description"`             // 描述
// }

// var targetModeMap = map[int32]bool{
// 	constdef.DefaultTargetMode: true,
// 	constdef.ConsulTargetMode:  true,
// }

// var methodMap = map[string]bool{
// 	http.MethodGet:     true,
// 	http.MethodHead:    true,
// 	http.MethodPost:    true,
// 	http.MethodPut:     true,
// 	http.MethodPatch:   true,
// 	http.MethodDelete:  true,
// 	http.MethodConnect: true,
// 	http.MethodOptions: true,
// 	http.MethodTrace:   true,
// }

// var loadBalanceMap = map[string]bool{
// 	constdef.RandLoadBalance:       true,
// 	constdef.RoundRobinLoadBalance: true,
// 	constdef.IPHashLoadBalance:     true,
// 	constdef.URLHashLoadBalance:    true,
// }

// var authMap = map[string]bool{
// 	constdef.KeyLess: true,
// 	constdef.AuthJWT: true,
// }

// type createAPIHandler struct {
// 	Ctx              *gin.Context
// 	Params           CreateAPIParams
// 	TargetURL        *url.URL
// 	IPWhiteList      []net.IP
// 	IPBlackList      []net.IP
// 	APIConfig        *dal.APIGatewayConfig
// 	APIConfigHistory *dal.APIGatewayConfigHistory
// }

// func buildCreateAPIHandler(c *gin.Context) *createAPIHandler {
// 	return &createAPIHandler{
// 		Ctx: c,
// 	}
// }

// type _StdResponse struct {
// 	Prompts string `json:"prompts"`
// 	Status  int32  `json:"status"`
// 	Message string `json:"message"`
// 	Data    string `json:"data"`
// }

// // @CreateAPI 创建API
// // @Description 创建API的接口
// // @Accept application/json
// // @Produce application/json
// // @Param object body CreateAPIParams true "请求body"
// // @Success 200 {object} _StdResponse
// // @Router /gateway-admin/api/create [post]
// func CreateAPI(c *gin.Context) (out *resp.JSONOutput) {

// 	h := buildCreateAPIHandler(c)

// 	err := h.CheckParams()
// 	if err != nil {
// 		logger.Error("[CreateAPI] params-error: err=%v", err)
// 		return resp.SampleJSON(c, resp.RespCodeParamsError, nil)
// 	}

// 	err = h.Process()
// 	if err != nil {
// 		logger.Error("[CreateAPI] Write API to DB failed: err=%v", err)
// 		return resp.SampleJSON(c, resp.RespDatabaseError, nil)
// 	}

// 	err = h.Notify()
// 	if err != nil {
// 		logger.Error("[CreateAPI] Notify failed: err=%v", err)
// 		return resp.SampleJSON(c, resp.RespCodeRedisError, nil)
// 	}
// 	return resp.SampleJSON(c, resp.RespCodeSuccess, nil)
// }

// func (h *createAPIHandler) CheckParams() (err error) {
// 	err = h.Ctx.Bind(&h.Params)
// 	if err != nil {
// 		logger.Error("[createAPIHandler-checkParams] params-err: err=%v", err)
// 		return err
// 	}
// 	if h.Params.Pattern[0] != '/' || h.Params.Pattern != path.Clean(h.Params.Pattern) {
// 		err = errors.New("http: pattern is invalid")
// 		logger.Error("[createAPIHandler-checkParams] params-err: pattern=%v", h.Params.Pattern)
// 		return err
// 	}

// 	if !methodMap[h.Params.Method] {
// 		err = errors.New("http: method is invalid")
// 		logger.Error("[createAPIHandler-checkParams] params-err: method=%v", h.Params.Method)
// 		return err
// 	}

// 	if !targetModeMap[h.Params.TargetMode] {
// 		err = errors.New("target_modo: mode is invalid")
// 		logger.Error("[createAPIHandler-checkParams] params-err: mode=%v", h.Params.TargetMode)
// 		return err
// 	}

// 	if h.Params.TargetMode == constdef.DefaultTargetMode {
// 		h.TargetURL, err = url.Parse(h.Params.TargetURL)
// 		if err != nil {
// 			logger.Error("[createAPIHandler-checkParams] params-err: mode=%v, target_url=%v", h.Params.TargetMode, h.Params.TargetURL)
// 			return err
// 		}
// 	} else if h.Params.TargetMode == constdef.ConsulTargetMode {
// 		if h.Params.TargetServiceName == "" {
// 			err = errors.New("consul: service_name is invalid")
// 			logger.Error("[createAPIHandler-checkParams] params-err: service_name=%v", h.Params.TargetServiceName)
// 			return err
// 		}
// 		h.Params.TargetLb = strings.ToUpper(h.Params.TargetLb)
// 		if h.Params.TargetLb == "" {
// 			h.Params.TargetLb = constdef.RandLoadBalance
// 		}
// 		if !loadBalanceMap[h.Params.TargetLb] {
// 			err = errors.New("consul: loadbalance is invalid")
// 			logger.Error("[createAPIHandler-checkParams] params-err: lb=%v", h.Params.TargetLb)
// 			return err
// 		}
// 	}

// 	h.Params.Auth = strings.ToUpper(h.Params.Auth)
// 	if h.Params.Auth == "" {
// 		h.Params.Auth = constdef.KeyLess
// 	}
// 	if !authMap[h.Params.Auth] {
// 		err = errors.New("auth: auth type is invalid")
// 		logger.Error("[createAPIHandler-checkParams] params-err: auth=%v", h.Params.Auth)
// 		return err
// 	}

// 	if h.Params.IPWhiteList != "" {
// 		h.IPWhiteList, err = tools.GetNetIPList(h.Params.IPWhiteList)
// 		if err != nil {
// 			logger.Error("[createAPIHandler-checkParams] params-err: ip_white_list=%v", h.Params.IPWhiteList)
// 			return err
// 		}
// 	}
// 	if h.Params.IPBlackList != "" {
// 		h.IPBlackList, err = tools.GetNetIPList(h.Params.IPBlackList)
// 		if err != nil {
// 			logger.Error("[createAPIHandler-checkParams] params-err: ip_black_list=%v", h.Params.IPBlackList)
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (h *createAPIHandler) Process() (err error) {

// 	// 开启数据库事务，只有下列操作全部通过，才往数据库里写
// 	db := storage.MysqlClient.Begin()
// 	defer func() {
// 		if err != nil {
// 			logger.Error("[createAPIHandler-Process] create api failed: err=%v", err)
// 			db.Rollback()
// 		} else {
// 			db.Commit()
// 			logger.Info("[createAPIHandler-Process] create api succeed")
// 		}
// 	}()

// 	// APIConfig: Write to Mysql
// 	h.packAPIConfig()
// 	err = dal.CreateAPI(db, h.APIConfig)
// 	if err != nil {
// 		logger.Error("[createAPIHandler-Process] write a api failed: err=%v", err)
// 		return err
// 	}

// 	// APIConfigHistory: Write to Mysql
// 	h.packAPIConfigHistory()
// 	err = dal.CreateAPIHistory(db, h.APIConfigHistory)
// 	if err != nil {
// 		logger.Error("[createAPIHandler-Process] write a api history failed: err=%v", err)
// 		return
// 	}

// 	// 增加API配置id到redis集合
// 	err = storage.RedisClient.SAdd(h.Ctx.Request.Context(), constdef.AllAPIConfigID, h.APIConfig.ID).Err()
// 	if err != nil {
// 		logger.Error("[createAPIHandler-Process] save api_id to redis set failed: err=%v", err)
// 		return err
// 	}

// 	// APIConfig: Write to Redis
// 	msgBytes, err := json.Marshal(h.APIConfig)
// 	if err != nil {
// 		logger.Error("[createAPIHandler-Process] marshal failed: err=%v", err)
// 		return err
// 	}
// 	key := fmt.Sprintf(constdef.APIConfigKeyFmt, h.APIConfig.ID)
// 	err = storage.RedisClient.Set(h.Ctx.Request.Context(), key, string(msgBytes), 3*30*24*time.Hour).Err()
// 	if err != nil {
// 		logger.Error("[createAPIHandler-Process] write to redis failed: err=%v", err)
// 		return err
// 	}
// 	return
// }

// func (h *createAPIHandler) Notify() (err error) {
// 	// Notify update
// 	for i := 0; i < 3; i++ {
// 		err = logic.RedisPub(h.Ctx.Request.Context(), constdef.UpdateGatewayRoute, "by create api")
// 		if err == nil {
// 			return
// 		}
// 		logger.Error("[createAPIHandler-Notify] Notify failed: try=%v, channl=%v, message=%v",
// 			i+1, constdef.UpdateGatewayRoute, "by create api")
// 	}
// 	return
// }

// func (h *createAPIHandler) packAPIConfig() {
// 	h.APIConfig = &dal.APIGatewayConfig{
// 		Pattern:           h.Params.Pattern,
// 		Method:            h.Params.Method,
// 		APIName:           h.Params.APIName,
// 		TargetMode:        h.Params.TargetMode,
// 		TargetTimeout:     h.Params.TargetTimeout,
// 		TargetStripPrefix: h.Params.TargetStripPrefix,
// 		MaxQPS:            h.Params.MaxQPS,
// 		Auth:              h.Params.Auth,
// 		Status:            1,
// 		Description:       h.Params.Description,
// 	}
// 	if h.Params.TargetMode == constdef.DefaultTargetMode {
// 		h.APIConfig.TargetHost = h.TargetURL.Host
// 		h.APIConfig.TargetScheme = h.TargetURL.Scheme
// 		h.APIConfig.TargetPath = h.TargetURL.Path
// 	} else if h.Params.TargetMode == constdef.ConsulTargetMode {
// 		h.APIConfig.TargetLb = h.Params.TargetLb
// 		h.APIConfig.TargetServiceName = h.Params.TargetServiceName
// 	}
// 	ipBlackList := make([]string, 0)
// 	for _, ip := range h.IPBlackList {
// 		ipBlackList = append(ipBlackList, ip.To4().String())
// 	}
// 	if len(ipBlackList) > 0 {
// 		h.APIConfig.IPBlackList = strings.Join(ipBlackList, ",")
// 	}
// 	ipWhiteList := make([]string, 0)
// 	for _, ip := range h.IPWhiteList {
// 		ipWhiteList = append(ipWhiteList, ip.To4().String())
// 	}
// 	if len(ipWhiteList) > 0 {
// 		h.APIConfig.IPWhiteList = strings.Join(ipWhiteList, ",")
// 	}
// }

// func (h *createAPIHandler) packAPIConfigHistory() {
// 	h.APIConfigHistory = &dal.APIGatewayConfigHistory{
// 		APIID:             h.APIConfig.ID,
// 		Pattern:           h.APIConfig.Pattern,
// 		Method:            h.APIConfig.Method,
// 		APIName:           h.APIConfig.APIName,
// 		TargetMode:        h.APIConfig.TargetMode,
// 		TargetHost:        h.APIConfig.TargetHost,
// 		TargetScheme:      h.APIConfig.TargetScheme,
// 		TargetPath:        h.APIConfig.TargetPath,
// 		TargetServiceName: h.APIConfig.TargetServiceName,
// 		TargetStripPrefix: h.APIConfig.TargetStripPrefix,
// 		TargetLb:          h.APIConfig.TargetLb,
// 		MaxQPS:            h.APIConfig.MaxQPS,
// 		Auth:              h.APIConfig.Auth,
// 		IPWhiteList:       h.APIConfig.IPWhiteList,
// 		IPBlackList:       h.APIConfig.IPBlackList,
// 		Description:       h.APIConfig.Description,
// 	}
// }
