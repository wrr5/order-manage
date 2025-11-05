package models

import (
	"time"

	"gorm.io/gorm"
)

type Shipment struct {
	gorm.Model
	OrderID         string         `gorm:"column:order_id;not null;size:50" json:"order_id"`                                                                       // 订单ID
	ShippingCompany string         `gorm:"column:shipping_company;not null;size:100" json:"shipping_company"`                                                      // 快递公司
	TrackingNumber  string         `gorm:"column:tracking_number;not null;size:100;unique" json:"tracking_number"`                                                 // 快递单号
	ShipmentStatus  ShipmentStatus `gorm:"column:shipment_status;type:enum('pending','shipped','delivered','cancelled');default:'pending'" json:"shipment_status"` // 发货状态
	ShipmentTime    *time.Time     `gorm:"column:shipment_time" json:"shipment_time"`                                                                              // 发货时间

	// 关联关系
	Order      Order       `gorm:"foreignKey:OrderID;references:OrderID" json:"order"`
	OrderItems []OrderItem `gorm:"foreignKey:ShipmentID;references:ID" json:"order_items"`
}

// 发货状态类型
type ShipmentStatus string

const (
	ShipmentPending   ShipmentStatus = "pending"
	ShipmentShipped   ShipmentStatus = "shipped"
	ShipmentDelivered ShipmentStatus = "delivered"
	ShipmentCancelled ShipmentStatus = "cancelled"
)
