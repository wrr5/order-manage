package models

// Zb 直播间表模型
type Zb struct {
	ZbID   string `gorm:"primaryKey;column:zbId;type:varchar(100);not null;comment:直播ID"`
	ZbName string `gorm:"column:zb_name;type:varchar(50);not null;comment:直播间名称"`
}

// TableName 设置表名
func (Zb) TableName() string {
	return "zb"
}
