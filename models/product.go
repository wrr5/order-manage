package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ProductID   string `gorm:"column:product_id;primaryKey;size:50" json:"product_id"`    // 商品ID
	ProductName string `gorm:"column:product_name;not null;size:200" json:"product_name"` // 商品名称

	// 关联关系
	OrderItems []OrderItem `gorm:"foreignKey:ProductID;references:ProductID" json:"order_items"`
}
