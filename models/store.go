package models

import (
	"time"
)

// Store 门店表模型
type Store struct {
	VzStoreID    string    `gorm:"primaryKey;column:vz_store_id;type:varchar(50);not null;comment:门店微赞ID"`
	StoreName    string    `gorm:"column:store_name;type:varchar(200);not null;index:idx_store_name;comment:门店名称"`
	Address      string    `gorm:"column:address;type:varchar(500);not null;comment:门店地址"`
	Receiver     string    `gorm:"column:receiver;type:varchar(50);comment:收货人"`
	ContactPhone string    `gorm:"column:contact_phone;type:varchar(20);index:idx_contact_phone;comment:联系方式"`
	CreatedTime  time.Time `gorm:"column:created_time;autoCreateTime;index:idx_created_time;comment:创建时间"`
	UpdatedTime  time.Time `gorm:"column:updated_time;autoUpdateTime;comment:更新时间"`
	Status       int8      `gorm:"column:status;type:tinyint;default:1;comment:状态（1正常，0停用）"`
}

// TableName 设置表名
func (Store) TableName() string {
	return "stores"
}

// 状态常量
const (
	StoreStatusActive   = 1 // 正常
	StoreStatusInactive = 0 // 停用
)
