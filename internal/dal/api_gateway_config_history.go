package dal

import (
	"time"

	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"gorm.io/gorm"
)

// APIGatewayConfigHistory ...
type APIGatewayConfigHistory struct {
	ID                int64     `gorm:"column:id" json:"id"`
	APIID             int64     `gorm:"column:api_id" json:"api_id"`
	Pattern           string    `gorm:"column:pattern" json:"pattern"`
	Method            string    `gorm:"column:method" json:"method"`
	APIName           string    `gorm:"column:api_name" json:"api_name"`
	TargetMode        int32     `gorm:"column:target_mode" json:"target_mode"`
	TargetHost        string    `gorm:"column:target_host" json:"target_host"`
	TargetScheme      string    `gorm:"column:target_scheme" json:"target_scheme"`
	TargetPath        string    `gorm:"column:target_path" json:"target_path"`
	TargetServiceName string    `gorm:"column:target_service_name" json:"target_service_name"`
	TargetLb          string    `gorm:"column:target_lb" json:"target_lb"`
	MaxQps            int32     `gorm:"column:max_qps" json:"max_qps"`
	Auth              string    `gorm:"column:auth" json:"auth"`
	IPWhiteList       string    `gorm:"column:ip_white_list" json:"ip_white_list"`
	IPBlackList       string    `gorm:"column:ip_black_list" json:"ip_black_list"`
	CreatedTime       time.Time `gorm:"column:created_time" json:"created_time"`
	ModifiedTime      time.Time `gorm:"column:modified_time" json:"modified_time"`
	Version           int32     `gorm:"column:version" json:"version"`
	Description       string    `gorm:"column:description" json:"description"`
}

func CreateAPIHistory(db *gorm.DB, apiConfigHistory *APIGatewayConfigHistory) error {
	db = db.Debug().Model(&APIGatewayConfigHistory{}).Create(apiConfigHistory)
	if db.Error != nil {
		logger.Error("[CreateAPIHistory] create API History failed: apiConfigHistory=%+v, err=%v", apiConfigHistory, db.Error)
		return db.Error
	}
	return nil
}

func GetAPIHistoryByApiID(db *gorm.DB, apiID int64) (history *APIGatewayConfigHistory, err error) {
	historyList := make([]*APIGatewayConfigHistory, 0)
	db = db.Debug().Model(&APIGatewayConfigHistory{}).Where("api_id = ?", apiID).Order("version DESC").Find(&historyList)
	if db.Error != nil {
		err = db.Error
		logger.Error("[GetAPIHistoryByApiID] select failed: api_id=%v, err=%v", apiID, err)
		return
	}
	if len(historyList) == 0 {
		return
	}
	history = historyList[0]
	return
}
