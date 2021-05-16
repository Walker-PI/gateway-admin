package resp

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// JSONOutput ...
type JSONOutput struct {
	context    *gin.Context
	HTTPStatus int
	Resp       interface{}
}

// NewJSONOutput ...
func NewJSONOutput(c *gin.Context, httpStatus int, rsp interface{}) *JSONOutput {
	return &JSONOutput{
		context:    c,
		HTTPStatus: httpStatus,
		Resp:       rsp,
	}
}

// SampleJSON ...
func SampleJSON(c *gin.Context, p ErrorCode, data interface{}) *JSONOutput {
	return NewJSONOutput(c, http.StatusOK, NewStdResponse(p, data))
}

// GetRespRawData ...
func (s *JSONOutput) GetRespRawData() []byte {
	vi := reflect.ValueOf(s.Resp)
	if vi.Kind() == reflect.Ptr && vi.IsNil() {
		return []byte("")
	}
	rawData, _ := json.Marshal(s.Resp)
	return rawData
}

// Write ...
func (s *JSONOutput) Write() {
	s.context.Writer.WriteHeader(s.HTTPStatus)
	s.context.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = s.context.Writer.Write(s.GetRespRawData())
}
