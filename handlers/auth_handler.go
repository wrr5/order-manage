package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/order-manage/global"
	"github.com/wrr5/order-manage/models"
	"github.com/wrr5/order-manage/services"
	"github.com/wrr5/order-manage/tools"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Login(c *gin.Context) {
	db := global.DB

	type loginRequest struct {
		PhoneNumber string `form:"phone_mumber" json:"phone_mumber" binding:"required,len=11"`
		RealName    string `form:"real_name" json:"real_name"`
		UserType    string `form:"user_type" json:"user_type"`
		PeriodZbid  string `form:"period_zbid" json:"period_zbid" binding:"required"`
		Password    string `form:"password" json:"password" binding:"max=20"`
	}
	var req loginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	var zb models.Zb
	var user models.User
	if err := db.First(&zb, "zbid = ?", req.PeriodZbid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "直播间不存在",
			})
		} else {
			log.Printf("查询直播间失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "系统错误，请稍后重试",
			})
		}
		return
	}
	// 门店员工登录处理
	if req.UserType == "2" {
		vzResp, err := services.ValidatePhoneName(req.PhoneNumber, req.RealName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err := db.Where("phone_number = ? AND period_zbid = ?", req.PhoneNumber, req.PeriodZbid).Preload("Zb").First(&user).Error; err != nil {
			user.PhoneNumber = req.PhoneNumber
			user.RealName = req.RealName
			user.PeriodZbid = &req.PeriodZbid
			value, err := strconv.ParseInt(req.UserType, 10, 8)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "用户类型不合法",
				})
				return
			}
			user.UserType = int8(value)
			result := db.Create(&user)
			if result.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "用户创建失败",
				})
				return
			}
			log.Printf("成功创建用户: %s", user.RealName)
		}
		var store models.Store
		store.VzStoreID = strconv.FormatInt(vzResp.DataObj.Records[0].StoreID, 10)
		store.StoreName = *vzResp.DataObj.Records[0].StoreName
		store.Address = *vzResp.DataObj.Records[0].StoreAddress

		// 使用OnConflict实现upsert
		result := global.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "vz_store_id"}},                                      // 冲突判断列
			DoUpdates: clause.AssignmentColumns([]string{"store_name", "address", "updated_time"}), // 更新字段
		}).Create(&store)

		if result.Error != nil {
			// 处理错误
			log.Printf("创建/更新门店失败: %v", result.Error)
			return
		}

		if result.RowsAffected > 0 {
			log.Printf("成功处理门店: %s", store.VzStoreID)

			// 成功创建门店后，更新门店员工的VzStoreID
			user.VzStoreID = &store.VzStoreID
			db.Save(&user)
		}
	} else {
		if err := db.Where("phone_number = ? AND period_zbid = ?", req.PhoneNumber, req.PeriodZbid).Preload("Zb").First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
			})
			return
		}
		// 验证密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
			})
			return
		}
	}

	token, err := tools.GenerateJWT(user.UserID, user.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成token失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "登录成功",
		"authorization": "Bearer " + token,
		"user":          user,
	})
}
