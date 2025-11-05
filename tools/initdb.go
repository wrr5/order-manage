package tools

import (
	"fmt"
	"log"

	"github.com/wrr5/general-management/config"
	"github.com/wrr5/general-management/global"
	"github.com/wrr5/general-management/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	cfg := config.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Name)
	db, error := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if error != nil {
		log.Fatal("数据库连接失败：", error)
	}

	// 自动迁移（如果表不存在则创建, 已存在则检查有无新增字段，不会修改字段名和删除字段）
	err := db.AutoMigrate(&models.User{}, &models.Store{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
	log.Println("数据库连接并迁移成功!")

	global.DB = db
	return db
}
