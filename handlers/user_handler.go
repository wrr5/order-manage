package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/general-management/global"
	"github.com/wrr5/general-management/models"
)

func ShowUserPage(c *gin.Context) {
	userInterface, _ := c.Get("user")
	user := userInterface.(models.User)
	db := global.DB

	var users []models.User
	err := db.Order("created_at DESC").Find(&users).Error
	if err != nil {
		// 处理错误
		c.HTML(http.StatusOK, "user.html", gin.H{
			"CurrentPath": c.Request.URL.Path,
			"error":       err.Error(),
			"users":       []models.User{},
		})
		return
	}
	c.HTML(http.StatusOK, "user.html", gin.H{
		"CurrentPath": c.Request.URL.Path,
		"user":        user,
		"users":       users,
	})
}

func CreateUser(c *gin.Context) {
	// userInterface, _ := c.Get("user")
	// user := userInterface.(models.User)
	// db := global.DB

	// if user.UserType != "admin" {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"error": "当前用户不是管理员",
	// 	})
	// 	return
	// }

	type CreateRequest struct {
		Phone    string `form:"phone" binding:"required,len=11"`
		Name     string `form:"name" binding:"required"`
		Password string `form:"password" binding:"required,min=6,max=20"`
		UserType string `form:"userType" binding:"required"`
	}
	var req CreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}
	fmt.Println(req)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建成功",
	})
}
