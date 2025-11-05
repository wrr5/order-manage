package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderID         string    `gorm:"column:order_id;primaryKey;size:50" json:"order_id"`                 // 订单ID
	BuyerName       string    `gorm:"column:buyer_name;not null;size:100" json:"buyer_name"`              // 买家名称
	BuyerPhone      string    `gorm:"column:buyer_phone;not null;size:20" json:"buyer_phone"`             // 电话
	ShippingAddress string    `gorm:"column:shipping_address;type:text;not null" json:"shipping_address"` // 收货地址
	OrderTime       time.Time `gorm:"column:order_time;not null" json:"order_time"`                       // 下单时间
	BuyerMessage    string    `gorm:"column:buyer_message;type:text" json:"buyer_message"`                // 买家留言

	// 关联关系
	OrderItems []OrderItem `gorm:"foreignKey:OrderID;references:OrderID" json:"order_items"`
	Shipments  []Shipment  `gorm:"foreignKey:OrderID;references:OrderID" json:"shipments"`
}
