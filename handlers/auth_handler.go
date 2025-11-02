package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/general-management/global"
	"github.com/wrr5/general-management/models"
	"github.com/wrr5/general-management/tools"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"CurrentURL": c.Request.URL.Path,
	})
}

func Login(c *gin.Context) {
	type loginUser struct {
		Mobile   string `form:"uname" binding:"required,len=11"`     // 手机号
		Password string `form:"pwd" binding:"required,min=6,max=20"` // 密码
	}

	var loginReq loginUser
	if err := c.ShouldBind(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	db := global.DB

	// 查找用户
	var user models.User
	if err := db.Where("mobile = ?", loginReq.Mobile).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "系统错误，请稍后重试",
			})
		}
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "密码错误",
		})
		return
	}

	// 更新最后登录时间
	db.Model(&user).Update("last_login_at", time.Now())

	// 生成JWT token
	token, err := tools.GenerateJWT(user.ID, user.Mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "登录失败，请重试",
		})
		return
	}
	// 设置 HTTP-only Cookie
	c.SetCookie("auth_token", "Bearer "+token, 3600*72, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功！",
	})
}

func Logout(c *gin.Context) {
	// 清除cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	// 返回成功信息或重定向到登录页
	// c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
	c.Redirect(http.StatusFound, "/login")
}

func ShowInform(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "通知列表",
	})
}
