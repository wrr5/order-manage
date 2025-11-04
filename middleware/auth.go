package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/general-management/global"
	"github.com/wrr5/general-management/models"
	"github.com/wrr5/general-management/tools"
)

// 获取当前用户信息
func GetUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 Cookie 获取 token
		token, err := c.Cookie("auth_token")
		if err == nil && len(token) > 7 {
			// 有 token，尝试解析
			if claims, err := tools.ParseJWT(token[7:]); err == nil {
				// token 有效，获取用户信息
				if userIDFloat, ok := claims["user_id"].(float64); ok {
					userID := uint(userIDFloat)
					var user models.User
					if global.DB.First(&user, userID).Error == nil {
						c.Set("user", user)
						c.Next()
						return
					}
				}
			}
		}
		// 没有有效 token 或解析失败，设置为空用户
		c.Set("user", models.User{})
		c.Next()
	}
}

// 验证登录状态的中间件，未登录则跳转到/login
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 Cookie 获取 token
		token, err := c.Cookie("auth_token")
		// 从Authorization Header获取
		// authHeader := c.GetHeader("Authorization")
		// if authHeader == "" {
		// 	c.JSON(401, gin.H{"error": "未提供认证信息"})
		// 	return
		// }

		if err != nil || len(token) <= 7 {
			// 没有token或token格式不对，跳转到登录页
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// 解析JWT token
		claims, err := tools.ParseJWT(token[7:]) // 去掉 "Bearer " 前缀
		if err != nil {
			// token无效，跳转到登录页
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 验证用户ID是否存在
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			// 用户ID格式错误，跳转到登录页
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		userID := uint(userIDFloat)
		var user models.User
		if err := global.DB.First(&user, userID).Error; err != nil {
			// 用户不存在，跳转到登录页
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// 用户验证通过，设置用户信息到上下文
		c.Set("user", user)
		c.Next()
	}
}
