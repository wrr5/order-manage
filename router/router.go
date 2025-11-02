package router

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/general-management/handlers"
	"github.com/wrr5/general-management/middleware"
)

// SetupRouter 配置所有路由
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.GetUserMiddleware())

	// 添加模板函数
	r.SetFuncMap(template.FuncMap{})

	// 注册文章相关路由
	setAuthRoutes(r)
	setInformRoutes(r)
	// 根路径跳转
	r.GET("/", handlers.ShowLogin)
	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "notfound.html", gin.H{"error": "页面不存在"})
	})

	return r
}

func setAuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.GET("/login", handlers.ShowLogin)
		auth.POST("/login", handlers.Login)
		auth.GET("/logout", handlers.Logout)
	}
}

func setInformRoutes(r *gin.Engine) {
	inform := r.Group("/inform")
	{
		inform.GET("", handlers.ShowInform)
	}
}
