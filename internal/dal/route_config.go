package dal

import (
	"time"

	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"gorm.io/gorm"
)

type RouteConfig struct {
	ID                   int64     `gorm:"column:id" json:"id"`
	GroupID              int64     `gorm:"column:group_id" json:"group_id"`
	GroupName            string    `gorm:"column:group_name" json:"group_name"`
	Source               string    `gorm:"column:source" json:"source"`
	Pattern              string    `gorm:"column:pattern" json:"pattern"`
	Methods              string    `gorm:"column:methods" json:"methods"`
	RateLimit            int32     `gorm:"column:rate_limit" json:"rate_limit"`
	AuthType             string    `gorm:"column:auth_type" json:"auth_type"`
	IPWhiteList          string    `gorm:"column:ip_white_list" json:"ip_white_list"`
	IPBlackList          string    `gorm:"column:ip_black_list" json:"ip_black_list"`
	TargetURL            string    `gorm:"column:target_url" json:"target_url"`
	TargetTimeout        int32     `gorm:"column:target_timeout" json:"target_timeout"`
	Discovery            string    `gorm:"column:discovery" json:"discovery"`
	DiscoveryPath        string    `gorm:"column:discovery_path" json:"discovery_path"`
	DiscoveryServiceName string    `gorm:"column:discovery_service_name" json:"discovery_service_name"`
	DiscoveryLoadBalance string    `gorm:"column:discovery_load_balance" json:"discovery_load_balance"`
	Deleted              int32     `gorm:"column:deleted" json:"deleted"`
	CreatedTime          time.Time `gorm:"column:created_time;default:CURRENT_TIMESTAMP" json:"created_time"`
	ModifiedTime         time.Time `gorm:"column:modified_time;default:CURRENT_TIMESTAMP" json:"modified_time"`
}

func CreateRouteConfig(db *gorm.DB, routeConfig *RouteConfig) error {
	db = db.Debug().Model(RouteConfig{}).Create(routeConfig)
	if db.Error != nil {
		logger.Error("[CreateRouteConfig] failed: err=%v", db.Error)
		return db.Error
	}
	return nil
}

func GetRouteConfigByID(routeID int64) (routeConfig *RouteConfig, err error) {
	routeConfigList := make([]*RouteConfig, 0)
	dbRes := storage.MysqlClient.Debug().Model(&RouteConfig{}).Where("id = ?", routeID).First(&routeConfigList)
	if dbRes.Error != nil {
		err = dbRes.Error
		logger.Error("[GetRouteConfigByID] failed: err=%v", err)
		return
	}
	if len(routeConfigList) > 0 {
		routeConfig = routeConfigList[0]
	}
	return
}

func UpdateRouteConfig(db *gorm.DB, routeID int64, routeConfig *RouteConfig) error {
	db = db.Debug().Model(RouteConfig{}).Where("id = ?", routeID).Updates(routeConfig)
	if db.Error != nil {
		logger.Error("[UpdateRouteConfig] failed: err=%v", db.Error)
		return db.Error
	}
	return nil
}

func DeleteRouteConfig(db *gorm.DB, routeID int64) error {
	db = db.Debug().Model(RouteConfig{}).Delete("id = ?", routeID)
	if db.Error != nil {
		logger.Error("[DeleteRouteConfig] failed: err=%v", db.Error)
		return db.Error
	}
	return nil
}
