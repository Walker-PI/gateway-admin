package handler

import (
	"fmt"
	"time"

	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/resp"
	"github.com/Walker-PI/gateway-admin/pkg/tools"
	"github.com/gin-gonic/gin"
)

const (
	secretKey = "Walker-PI"
)

type TokenParams struct {
	SecretKey string `form:"secret_key" json:"secret_key" binding:"required"`
	UserID    int64  `form:"user_id" json:"user_id" binding:"required"`
}

type getTokenHandler struct {
	Ctx    *gin.Context
	Params TokenParams
	Resp   map[string]string
}

func buildGetTokenHandler(c *gin.Context) *getTokenHandler {
	return &getTokenHandler{
		Ctx: c,
	}
}

func GetToken(c *gin.Context) (out *resp.JSONOutput) {
	h := buildGetTokenHandler(c)
	err := h.CheckParams()
	if err != nil {
		logger.Error("[GetToken] CheckParams failed: err=%v", err)
		return resp.SampleJSON(c, resp.RespCodeParamsError, nil)
	}

	err = h.Process()
	if err != nil {
		logger.Error("[GetToken] get token failed: err=%v", err)
		return resp.SampleJSON(c, resp.RespCodeServerException, nil)
	}
	return resp.SampleJSON(c, resp.RespCodeSuccess, h.Resp)
}

func (h *getTokenHandler) CheckParams() error {
	err := h.Ctx.Bind(&h.Params)
	if err != nil {
		logger.Error("[getTokenHandler-CheckParams] params-err: err=%v", err)
		return err
	}
	if h.Params.SecretKey != secretKey {
		logger.Error("[getTokenHandler-CheckParams] secretKey is invalid: secretKey=%v", secretKey)
		return fmt.Errorf("secret_key is invalid")
	}
	return nil
}

func (h *getTokenHandler) Process() error {

	token, err := tools.GenerateToken(h.Params.UserID, 30*time.Minute)
	if err != nil {
		return err
	}
	h.Resp = make(map[string]string)
	h.Resp["token"] = token
	return nil
}
