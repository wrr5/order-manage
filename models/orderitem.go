package models

import "gorm.io/gorm"

type OrderItem struct {
	gorm.Model
	OrderID       string `gorm:"column:order_id;not null;size:50" json:"order_id"`     // 订单ID
	ProductID     string `gorm:"column:product_id;not null;size:50" json:"product_id"` // 商品ID
	Specification string `gorm:"column:specification;size:200" json:"specification"`   // 规格
	Quantity      int    `gorm:"column:quantity;not null;default:1" json:"quantity"`   // 商品数量
	ShipmentID    *uint  `gorm:"column:shipment_id" json:"shipment_id"`                // 所属包裹ID

	// 关联关系
	Order    Order    `gorm:"foreignKey:OrderID;references:OrderID" json:"order"`
	Product  Product  `gorm:"foreignKey:ProductID;references:ProductID" json:"product"`
	Shipment Shipment `gorm:"foreignKey:ShipmentID;references:ID" json:"shipment"`
}
