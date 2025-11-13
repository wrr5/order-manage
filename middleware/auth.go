package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/wrr5/order-manage/global"
	"github.com/wrr5/order-manage/models"
	"github.com/wrr5/order-manage/tools"
)

// 验证登录状态的中间件，未登录则跳转到/login
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// token, err := c.Cookie("auth_token")
		// 从Authorization Header获取
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "未提供认证信息"})
			c.Abort()
			return
		}

		if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "Authorization header 格式应为: Bearer <token>"})
			c.Abort()
			return
		}

		// 去掉 "Bearer " 前缀
		tokenString := authHeader[7:]

		// 解析JWT token
		claims, err := tools.ParseJWT(tokenString)
		if err != nil {
			// token无效，跳转到登录页
			c.JSON(401, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		// 验证用户ID是否存在
		userID := claims.UserID
		var user models.User
		if err := global.DB.First(&user, userID).Error; err != nil {
			// 用户不存在，跳转到登录页
			c.JSON(401, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		// 用户验证通过，设置用户信息到上下文
		c.Set("user", user)
		c.Next()
	}
}
