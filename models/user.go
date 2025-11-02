package models

import (
	"gorm.io/gorm"
)

// UserType 用户类型枚举
type UserType string

const (
	UserTypeShopOwner       UserType = "shop_owner"       // 店长
	UserTypeFactory         UserType = "factory"          // 工厂
	UserTypeCustomerService UserType = "customer_service" // 客服
	UserTypePlatformOwner   UserType = "platform_owner"   // 平台老板
)

type User struct {
	gorm.Model
	Mobile   string   `gorm:"type:varchar(11);uniqueIndex;not null" json:"mobile"`
	Password string   `gorm:"type:varchar(255);not null" json:"-"`
	UserType UserType `gorm:"type:varchar(20);not null" json:"user_type" binding:"required"`
	Level    uint     `gorm:"type:tinyint;default:null" json:"level"`

	PlatformOwnerID *uint `gorm:"index;default:null;constraint:OnDelete:SET NULL" json:"platform_owner_id"`
	PlatformOwner   *User `gorm:"foreignKey:PlatformOwnerID" json:"platform_owner,omitempty"`
}
