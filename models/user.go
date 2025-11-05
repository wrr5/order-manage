package models

import (
	"gorm.io/gorm"
)

// UserType 用户类型枚举
type UserType string

const (
	UserTypeAdmin                  UserType = "admin"                    // 管理员
	UserTypePlatformOwner          UserType = "platform_owner"           // 平台老板
	UserTypeShopOwner              UserType = "shop_owner"               // 店长
	UserTypeShopEmployee           UserType = "shop_employee"            // 门店员工
	UserTypeCustomerServiceManager UserType = "customer_service_manager" // 客服主管
	UserTypeCustomerService        UserType = "customer_service"         // 客服
	UserTypeFactory                UserType = "factory"                  // 工厂
)

type User struct {
	gorm.Model

	Phone    string   `gorm:"type:varchar(11);uniqueIndex;not null" json:"phone"`
	RealName string   `gorm:"type:varchar(20);not null" json:"RealName"`
	Password string   `gorm:"type:varchar(255);default:null" json:"-"`
	UserType UserType `gorm:"type:varchar(20);not null" json:"user_type" binding:"required"`
	Level    uint     `gorm:"type:tinyint;default:null" json:"level"`

	// 平台老板关联 - 多对一
	PlatformOwnerID *uint `gorm:"index;default:null" json:"platform_owner_id"`
	PlatformOwner   *User `gorm:"foreignKey:PlatformOwnerID;constraint:OnDelete:SET NULL" json:"platform_owner,omitempty"`

	// 店铺关联 - 多对一
	StoreID *uint  `gorm:"index;default:null" json:"store_id"`
	Store   *Store `gorm:"foreignKey:StoreID;constraint:OnDelete:SET NULL" json:"store,omitempty"`

	// 客服关联 - 多对一
	CustomerServiceID *uint `gorm:"index;default:null" json:"customer_service_id"`
	CustomerService   *User `gorm:"foreignKey:CustomerServiceID;constraint:OnDelete:SET NULL" json:"customer_service,omitempty"`
}
