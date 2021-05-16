package resp

type ErrorCode int32

const (
	RespCodeSuccess         ErrorCode = 0
	RespCodeParamsError     ErrorCode = 4001
	RespCodeUserExsit       ErrorCode = 4002
	RespCodeServerException ErrorCode = 5000
	RespDatabaseError       ErrorCode = 5001
	RespCodeRedisError      ErrorCode = 5002
	RespCodeRPCError        ErrorCode = 5003
)

type IErrorCode interface {
	Prompts() string
	Message() string
	Status() int32
}

func (p ErrorCode) Prompts() string {
	switch p {
	case RespCodeSuccess:
		return ""
	case RespCodeParamsError:
		return "请求参数错误"
	case RespCodeUserExsit:
		return "用户名已存在"
	case RespCodeServerException, RespDatabaseError,
		RespCodeRedisError, RespCodeRPCError:
		return "服务器内部错误，请稍后重试"
	}
	return "unkown error"
}

func (p ErrorCode) Message() string {
	switch p {
	case RespCodeSuccess:
		return "success"
	case RespCodeParamsError:
		return "params error"
	case RespCodeUserExsit:
		return "username exsited"
	case RespCodeServerException, RespDatabaseError,
		RespCodeRedisError, RespCodeRPCError:
		return "server exception"
	}
	return "unkown error"
}

func (p ErrorCode) Status() int32 {
	return int32(p)
}
