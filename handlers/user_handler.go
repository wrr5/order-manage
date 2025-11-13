package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/order-manage/global"
	"github.com/wrr5/order-manage/models"
	"github.com/wrr5/order-manage/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateUser(c *gin.Context) {
	db := global.DB

	type createRequest struct {
		PhoneNumber string `form:"phone_mumber" json:"phone_mumber" binding:"required,len=11"`
		RealName    string `form:"real_name" json:"real_name" binding:"required"`
		UserType    string `form:"user_type" json:"user_type" binding:"required"`
		PeriodZbid  string `form:"period_zbid" json:"period_zbid" binding:"required"`
		Password    string `form:"password" json:"password" binding:"required,min=6,max=20"`
		VzStoreID   string `form:"vz_store_id" json:"vz_store_id"`
		VzFactoryID string `form:"vz_factory_id" json:"vz_factory_id"`
	}

	var req createRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	var existingUser models.User
	if err := db.Where("phone_number = ? AND period_zbid = ?", req.PhoneNumber, req.PeriodZbid).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "用户已存在",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询用户失败",
		})
		return
	}

	if req.UserType == "2" {
		_, err := services.ValidatePhoneName(req.PhoneNumber, req.RealName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "门店用户信息验证失败",
			})
			return
		}
	}
	value, err := strconv.ParseInt(req.UserType, 10, 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户类型不合法",
		})
		return
	}
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "密码加密失败",
		})
		return
	}
	user := models.User{
		PhoneNumber: req.PhoneNumber,
		RealName:    req.RealName,
		UserType:    int8(value),
		PeriodZbid:  &req.PeriodZbid,
		Password:    string(hashedPassword),
	}

	if req.VzStoreID != "" {
		user.VzStoreID = &req.VzStoreID
	}

	if req.VzFactoryID != "" {
		user.VzFactoryID = &req.VzFactoryID
	}

	if result := db.Create(&user).Error; result != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户创建失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建用户成功",
		"user":    user,
	})
}

func GetUsers(c *gin.Context) {
	type getRequest struct {
		UserType string `form:"user_type" json:"user_type"`
	}
	var req getRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}
	db := global.DB
	var users []models.User
	if req.UserType != "" {
		if err := db.Where("user_type = ?", req.UserType).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查询用户失败",
			})
			return
		}
	} else {
		if err := db.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查询用户失败",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "查询用户列表成功",
		"data":    users,
		"total":   len(users),
	})
}

func GetUser(c *gin.Context) {

}
