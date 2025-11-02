package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	// JWT      JWTConfig
	Page PageConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type ServerConfig struct {
	Port  string
	Debug bool
}

// type JWTConfig struct {
// 	Secret string
// }

type PageConfig struct {
	Size string
}

var AppConfig *Config

func Init() {
	// 设置配置文件路径（可选，默认当前目录）
	viper.AddConfigPath(".")
	// 设置配置文件名称（不需要扩展名）
	viper.SetConfigName("config")
	// 设置配置文件类型
	viper.SetConfigType("yaml")

	// 读取环境变量（可选，Viper 可以自动读取环境变量）
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 将配置信息解析到结构体
	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("无法解析结构体: %v", err)
	}

	log.Println("配置加载成功！")
}
