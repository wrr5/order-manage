package models

import (
	"time"
)

// User 用户表模型
type User struct {
	UserID      int       `gorm:"primaryKey;column:user_id;autoIncrement;comment:用户ID"`
	UserType    int8      `gorm:"column:user_type;not null;comment:用户类型（1客服，2门店员工，3工厂员工）"`
	RealName    string    `gorm:"column:real_name;type:varchar(50);not null;comment:真实姓名"`
	PhoneNumber string    `gorm:"column:phone_number;type:varchar(20);not null;uniqueIndex:phone_number;index:idx_phone;comment:手机号"`
	Password    string    `gorm:"column:password;type:varchar(255);not null;comment:密码"`
	VzStoreID   *string   `gorm:"column:vz_store_id;type:varchar(50);index:idx_store;comment:门店ID"`
	VzFactoryID *string   `gorm:"column:vz_factory_id;type:varchar(50);index:idx_factory;comment:工厂ID"`
	PeriodZbid  *string   `gorm:"column:period_zbid;type:varchar(100);comment:期数"`
	CreatedTime time.Time `gorm:"column:created_time;autoCreateTime;comment:创建时间"`
	UpdatedTime time.Time `gorm:"column:updated_time;autoUpdateTime;comment:更新时间"`

	// 关联关系（可选，根据实际需要添加）
	Store Store `gorm:"foreignKey:VzStoreID;references:VzStoreID"`
	// Factory Factory `gorm:"foreignKey:VzFactoryID;references:VzFactoryID"`
	Zb Zb `gorm:"foreignKey:PeriodZbid;references:ZbID"`
}

// TableName 设置表名
func (User) TableName() string {
	return "user"
}

// 用户类型常量
const (
	UserTypeCustomerService = 1 // 客服
	UserTypeStoreEmployee   = 2 // 门店员工
	UserTypeFactoryEmployee = 3 // 工厂员工
)
