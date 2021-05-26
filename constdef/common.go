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

// Redis Pub channel
const (
	UpdateGatewayRoute = "update-gateway-route"
)

// Redis key
const (
	AllRouteConfigIDFmt = "all-route-config-id:source:%s"
	RouteConfigKeyFmt   = "route-config:api_id:%d:source:%s"
)

// source
const (
	SourceCloud = "CLOUD"
	SourceEdgex = "EDGEX"
)

// Discovery
const (
	DiscoveryEureka = "EUREKA"
	DiscoveryConsul = "CONSUL"
)
