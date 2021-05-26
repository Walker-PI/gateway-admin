package dal

import (
	"time"

	"github.com/Walker-PI/gateway-admin/pkg/logger"
	"github.com/Walker-PI/gateway-admin/pkg/storage"
	"gorm.io/gorm"
)

type RouteGroup struct {
	ID           int64     `gorm:"column:id" json:"id"`
	GroupName    string    `gorm:"column:group_name" json:"group_name"`
	Source       string    `gorm:"column:source" json:"source"`
	Status       int32     `gorm:"column:status" json:"status"`
	Deleted      int32     `gorm:"column:deleted" json:"deleted"`
	CreatedTime  time.Time `gorm:"column:created_time;default:CURRENT_TIMESTAMP" json:"created_time"`
	ModifiedTime time.Time `gorm:"column:modified_time;default:CURRENT_TIMESTAMP" json:"modified_time"`
	Description  string    `gorm:"column:description" json:"description"`
}

func CreateGroup(db *gorm.DB, routeGroup *RouteGroup) error {
	db = db.Debug().Model(RouteGroup{}).Create(routeGroup)
	if db.Error != nil {
		logger.Error("[CreateGroup] create group failed: routeGroup=%+v, err=%v", routeGroup, db.Error)
		return db.Error
	}
	return nil
}

func GetRouteGroupByID(id int64) (routeGroup *RouteGroup, err error) {
	if id == 0 {
		return
	}
	routeGroupList := make([]*RouteGroup, 0)
	dbRes := storage.MysqlClient.Debug().Model(&RouteGroup{}).Where("id = ?", id).First(&routeGroupList)
	if dbRes.Error != nil {
		err = dbRes.Error
		logger.Error("[GetRouteGroupByID] get routeGroup failed: err=%v", err)
		return
	}
	if len(routeGroupList) == 0 {
		return
	}
	routeGroup = routeGroupList[0]
	return
}
func UpdateRouteGroup(db *gorm.DB, id int64, routeGroup *RouteGroup) error {
	db = db.Debug().Model(RouteGroup{}).Where("id = ?", id).Updates(routeGroup)
	if db.Error != nil {
		logger.Error("[UpdateRouteGroup] failed: err=%v", db.Error)
		return db.Error
	}
	return nil
}
