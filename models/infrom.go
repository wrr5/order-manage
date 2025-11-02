package models

import (
	"gorm.io/gorm"
)

type Inform struct {
	gorm.Model
	InformType int    `gorm:"type:tinyint;not null" json:"inform_type"` // 通知类型 1:运营 2:售后 3:财务
	Content    string `gorm:"type:text;default:NULL" json:"content"`    // 通知内容
}
