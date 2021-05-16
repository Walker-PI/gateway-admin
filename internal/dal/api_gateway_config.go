package dal

import (
	"time"

	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"gorm.io/gorm"
)

// APIGatewayConfig ...
type APIGatewayConfig struct {
	ID                int64     `gorm:"column:id" json:"id"`
	Pattern           string    `gorm:"column:pattern" json:"pattern"`
	Method            string    `gorm:"column:method" json:"method"`
	APIName           string    `gorm:"column:api_name" json:"api_name"`
	TargetMode        int32     `gorm:"column:target_mode" json:"target_mode"`
	TargetHost        string    `gorm:"column:target_host" json:"target_host"`
	TargetScheme      string    `gorm:"column:target_scheme" json:"target_scheme"`
	TargetPath        string    `gorm:"column:target_path" json:"target_path"`
	TargetServiceName string    `gorm:"column:target_service_name" json:"target_service_name"`
	TargetLb          string    `gorm:"column:target_lb" json:"target_lb"`
	TargetTimeout     int64     `gorm:"column:target_timeout" json:"target_timeout"`
	MaxQps            int32     `gorm:"column:max_qps" json:"max_qps"`
	Auth              string    `gorm:"column:auth" json:"auth"`
	IPWhiteList       string    `gorm:"column:ip_white_list" json:"ip_white_list"`
	IPBlackList       string    `gorm:"column:ip_black_list" json:"ip_black_list"`
	CreatedTime       time.Time `gorm:"column:created_time" json:"created_time"`
	ModifiedTime      time.Time `gorm:"column:modified_time" json:"modified_time"`
	Deleted           int32     `gorm:"column:deleted" json:"deleted"`
	Description       string    `gorm:"column:description" json:"description"`
}

func CreateAPI(db *gorm.DB, apiConfig *APIGatewayConfig) error {
	db = db.Debug().Model(APIGatewayConfig{}).Create(apiConfig)
	if db.Error != nil {
		logger.Error("[CreateAPI] create api failed: apiConfig=%+v, err=%v", apiConfig, db.Error)
		return db.Error
	}
	return nil
}

func UpdateAPI(db *gorm.DB, id int64, apiConfig *APIGatewayConfig) error {
	db = db.Debug().Model(APIGatewayConfig{}).Where("id = ?", id).Updates(apiConfig)
	if db.Error != nil {
		logger.Error("[UpdateAPI] update api failed: apiConfig=%+v, err=%v", apiConfig, db.Error)
		return db.Error
	}
	return nil
}
