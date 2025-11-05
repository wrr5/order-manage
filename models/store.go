package models

import (
	"gorm.io/gorm"
)

type Store struct {
	gorm.Model
	StoreId string `gorm:"type:varchar(50);not null" json:"store_id"`
	Name    string `gorm:"type:varchar(50);not null" json:"name"`
	Address string `gorm:"type:varchar(255)" json:"address"`
}
