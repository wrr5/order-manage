package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/general-management/global"
	"github.com/wrr5/general-management/models"
	"github.com/wrr5/general-management/services"
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
		Phone    string `form:"phone" binding:"required,len=11"`          // 手机号
		Password string `form:"password" binding:"required,min=6,max=20"` // 密码
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
	if err := db.Where("phone = ?", loginReq.Phone).First(&user).Error; err != nil {
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

	// 生成JWT token
	token, err := tools.GenerateJWT(user.ID, user.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "登录失败，请重试",
		})
		return
	}
	// 设置 HTTP-only Cookie
	c.SetCookie("auth_token", "Bearer "+token, 3600*72, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "登录成功",
		"authorization": "Bearer " + token,
		"user":          user,
	})
	// c.Redirect(http.StatusFound, "/")
}

func Logout(c *gin.Context) {
	// 清除cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	// 返回成功信息或重定向到登录页
	// c.JSON(http.StatusOK, gin.H{
	// 	"success": true,
	// 	"message": "退出成功",
	// })
	c.Redirect(http.StatusFound, "/auth/login")
}

func ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func Register(c *gin.Context) {
	type RegisterRequest struct {
		Phone    string `form:"phone" binding:"required,len=11"`
		Name     string `form:"name" binding:"required"`
		Password string `form:"password" binding:"required,min=6,max=20"`
	}
	db := global.DB
	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}
	var apiReq services.ApiResponse
	if ok, apiResponse, err := services.ValidatePhoneName(req.Phone, req.Name); !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	} else {
		apiReq = apiResponse
	}

	var count int64
	// 检查手机号是否已存在
	db.Model(&models.User{}).Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "手机号已注册",
		})
		return
	}

	// 转换请求为User模型
	var user models.User
	user.Phone = req.Phone
	user.Name = req.Name
	switch apiReq.DataObj.Records[0].Level {
	case 1:
		user.UserType = models.UserTypeShopEmployee
		user.Level = 1

		var platformOwner models.User
		result := db.Where("phone = ?", apiReq.DataObj.Records[0].SupPhone).
			FirstOrCreate(&platformOwner, models.User{
				Phone:    apiReq.DataObj.Records[0].SupPhone,
				Name:     apiReq.DataObj.Records[0].SupName,
				UserType: models.UserTypePlatformOwner,
				Level:    3,
			})

		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "查找或创建平台老板失败: " + result.Error.Error(),
			})
			return
		} else {
			user.PlatformOwnerID = &platformOwner.ID
		}
	case 2:
		user.UserType = models.UserTypeShopOwner
		user.Level = 2

		var platformOwner models.User
		result := db.Where("phone = ?", apiReq.DataObj.Records[0].ParentPhone).
			FirstOrCreate(&platformOwner, models.User{
				Phone:    apiReq.DataObj.Records[0].ParentPhone,
				Name:     apiReq.DataObj.Records[0].ParentName,
				UserType: models.UserTypePlatformOwner,
				Level:    3,
			})

		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "查找或创建平台老板失败: " + result.Error.Error(),
			})
			return
		} else {
			user.PlatformOwnerID = &platformOwner.ID
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "当前用户不是员工或店长",
		})
		return
	}

	// 保存门店信息
	var shop models.Shop
	result := db.Where("name = ?", apiReq.DataObj.Records[0].StoreName).
		FirstOrCreate(&shop, models.Shop{
			Name:    *apiReq.DataObj.Records[0].StoreName,
			Address: *apiReq.DataObj.Records[0].StoreAddress,
		})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查找或创建门店失败: " + result.Error.Error(),
		})
		return
	} else {
		user.ShopID = &shop.ID
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "密码加密失败",
		})
		return
	}
	user.Password = string(hashedPassword)

	// 创建用户
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "注册失败",
			"db_err": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "注册成功！",
		"user":    user,
	})
}
