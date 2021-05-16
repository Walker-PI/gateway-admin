package constdef

const (
	TimeFormat = "2006-01-02 15:04:05"
)

// LoadBalance
const (
	RandLoadBalance       = "RAND"
	IPHashLoadBalance     = "IP_HASH"
	URLHashLoadBalance    = "URL_HASH"
	RoundRobinLoadBalance = "ROUND_ROBIN"
)

// Auth
const (
	KeyLess = "KEYLESS"
	AuthJWT = "JWT"
)

// Target Mode
const (
	DefaultTargetMode int32 = 1
	ConsulTargetMode  int32 = 2
)

// Redis Pub channel
const (
	UpdateGatewayRoute = "update-gateway-route"
)
