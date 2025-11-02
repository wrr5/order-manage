package main

import (
	"github.com/gin-gonic/gin"
	// "github.com/thinkerou/favicon"
	"github.com/wrr5/general-management/config"
	"github.com/wrr5/general-management/router"
	"github.com/wrr5/general-management/tools"
)

func main() {
	config.Init()
	if config.AppConfig.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	tools.InitDB()

	r := router.SetupRouter()

	r.LoadHTMLGlob("templates/**/*.html")
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")
	// r.Use(favicon.New("./static/images/favicon.ico"))

	r.Run(":" + config.AppConfig.Server.Port)
}
