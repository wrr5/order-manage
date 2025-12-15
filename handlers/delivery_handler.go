package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/order-manage/global"
	"github.com/wrr5/order-manage/models"
	"github.com/wrr5/order-manage/services"
	"github.com/wrr5/order-manage/tools"
)

func GetLogistics(c *gin.Context) {
	type queryExpressRequest struct {
		VzStoreID string `json:"VzStoreID" binding:"required"`
		StateText string `json:"stateText"`
	}
	var req queryExpressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	type ExpressResponse struct {
		ExpressNumber  string         `json:"ExpressNumber"`
		ActualQuantity int            `json:"ActualQuantity"`
		DeliveryTime   time.Time      `json:"deliveryTime"` // 发货时间
		LatestTrace    services.Trace `json:"latestTrace"`
	}
	type OrderResponse struct {
		VzOrderID         int               `json:"VzOrderID"`
		ExpressResponse   []ExpressResponse `json:"deliveries"`
		ProductID         string            `json:"productID"`
		ProductName       string            `json:"productName"`
		Accepted          int               `json:"Accepted"`
		UnAccepted        int               `json:"UnAccepted"`
		ActualQuantitySum int               `json:"ActualQuantitySum"`
		ProductSum        int               `json:"ProductSum"`
	}

	var Orders []OrderResponse

	db := global.DB
	var orders []models.Order
	var express []models.Express
	unAcceptedtotle := 0
	totle := 0
	db.Where("vz_store_id = ?", req.VzStoreID).Preload("Product").Find(&orders)
	for _, order := range orders {
		newOrder := OrderResponse{
			VzOrderID:   order.VzOrderID,
			ProductID:   order.VzProductID,
			ProductName: order.Product.ProductName,
			ProductSum:  order.Quantity * order.Unit,
		}
		db.Where("vz_order_id = ?", order.VzOrderID).Find(&express)
		actualQuantitySum := 0
		accepted := 0
		unAccepted := 0
		for _, exp := range express {
			actualQuantitySum += exp.ActualQuantity
			switch exp.IsAccepted {
			case 0:
				unAccepted += 1
			case 1:
				accepted += 1
			}
			logisticsResponse, err := services.QueryDelivery(exp.ExpressNumber)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}
			if req.StateText == "" {
				newExp := ExpressResponse{
					ExpressNumber:  exp.ExpressNumber,
					ActualQuantity: exp.ActualQuantity,
					DeliveryTime:   exp.CreatedTime,
					LatestTrace:    logisticsResponse.DataObj.LogisticsInfo.Traces[0],
				}
				newOrder.ExpressResponse = append(newOrder.ExpressResponse, newExp)
			} else {
				if logisticsResponse.DataObj.LogisticsInfo.StateText == req.StateText {
					newExp := ExpressResponse{
						ExpressNumber: exp.ExpressNumber,
						DeliveryTime:  exp.CreatedTime,
						LatestTrace:   logisticsResponse.DataObj.LogisticsInfo.Traces[0],
					}
					newOrder.ExpressResponse = append(newOrder.ExpressResponse, newExp)
				}
			}
		}
		newOrder.ActualQuantitySum = actualQuantitySum
		newOrder.Accepted = accepted
		newOrder.UnAccepted = unAccepted
		unAcceptedtotle += unAccepted
		totle += unAccepted + accepted
		Orders = append(Orders, newOrder)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "查询物流成功",
		"totle":           totle,
		"unAcceptedtotle": unAcceptedtotle,
		"data":            Orders,
	})
}

func GetDeliveryByProductName(c *gin.Context) {
	type queryExpressRequest struct {
		VzStoreID   string `json:"VzStoreID" binding:"required"`
		ProductName string `json:"ProductName"`
	}
	var req queryExpressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	type ExpressResponse struct {
		ExpressNumber string           `json:"ExpressNumber"`
		DeliveryTime  time.Time        `json:"deliveryTime"` // 发货时间
		Trace         []services.Trace `json:"latestTrace"`
	}
	type OrderResponse struct {
		VzOrderID       int               `json:"VzOrderID"`
		ExpressResponse []ExpressResponse `json:"deliveries"`
		ProductID       string            `json:"productID"`
		ProductName     string            `json:"productName"`
		ProductSum      int               `json:"ProductSum"`
	}

	var respOrders []OrderResponse
	db := global.DB
	var orders []models.Order
	if req.ProductName != "" {
		var productIDs []string
		db.Model(&models.Product{}).Where("product_name LIKE ?", "%"+req.ProductName+"%").Pluck("vz_product_id", &productIDs)
		db.Where("vz_store_id = ? AND vz_product_id IN (?)", req.VzStoreID, productIDs).
			Preload("Product").Preload("Expresses").Find(&orders)

		for _, order := range orders {
			newOrder := OrderResponse{
				VzOrderID:   order.VzOrderID,
				ProductID:   order.VzProductID,
				ProductName: order.Product.ProductName,
				ProductSum:  order.Quantity * order.Unit,
			}

			for _, exp := range order.Expresses {
				logisticsResponse, err := services.QueryDelivery(exp.ExpressNumber)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": err.Error(),
					})
					return
				}

				newExp := ExpressResponse{
					ExpressNumber: exp.ExpressNumber,
					DeliveryTime:  exp.CreatedTime,
					Trace:         logisticsResponse.DataObj.LogisticsInfo.Traces,
				}
				newOrder.ExpressResponse = append(newOrder.ExpressResponse, newExp)
			}
			respOrders = append(respOrders, newOrder)
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "按商品名筛选成功",
			"data":    respOrders,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "未传递商品名",
		"data":    respOrders,
	})
}

func GetDeliveryByProductId(c *gin.Context) {
	type queryExpressRequest struct {
		VzStoreID   string `json:"VzStoreID" binding:"required"`
		VzProductID string `json:"VzProductID"  binding:"required"`
	}
	var req queryExpressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	type ExpressResponse struct {
		ExpressNumber  string `json:"ExpressNumber"`
		ActualQuantity int    `json:"ActualQuantity"`
		IsAccepted     int8   `json:"IsAccepted"`
	}
	type OrderResponse struct {
		VzOrderID       int               `json:"VzOrderID"`
		ExpressResponse []ExpressResponse `json:"deliveries"`
		ProductSum      int               `json:"ProductSum"`
	}

	var respOrders []OrderResponse
	db := global.DB
	var orders []models.Order
	if req.VzProductID != "" {
		db.Where("vz_store_id = ? AND vz_product_id = ?", req.VzStoreID, req.VzProductID).Preload("Product").Preload("Expresses").Find(&orders)

		for _, order := range orders {
			newOrder := OrderResponse{
				VzOrderID:  order.VzOrderID,
				ProductSum: order.Quantity * order.Unit,
			}

			for _, exp := range order.Expresses {

				newExp := ExpressResponse{
					ExpressNumber:  exp.ExpressNumber,
					ActualQuantity: exp.ActualQuantity,
					IsAccepted:     exp.IsAccepted,
				}
				newOrder.ExpressResponse = append(newOrder.ExpressResponse, newExp)
			}
			respOrders = append(respOrders, newOrder)
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "按商品ID筛选成功",
			"data":    respOrders,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "未传递商品ID",
		"data":    respOrders,
	})
}

func GetAccepted(c *gin.Context) {
	type queryExpressRequest struct {
		VzStoreID string `json:"VzStoreID" binding:"required"`
		Accepted  string `json:"Accepted"`
	}
	var req queryExpressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	type ExpressResponse struct {
		ExpressNumber string           `json:"ExpressNumber"`
		IsAccepted    int8             `json:"IsAccepted"`
		DeliveryTime  time.Time        `json:"deliveryTime"` // 发货时间
		Trace         []services.Trace `json:"Trace"`
	}
	type OrderResponse struct {
		VzOrderID       int               `json:"VzOrderID"`
		ExpressResponse []ExpressResponse `json:"deliveries"`
		ProductID       string            `json:"productID"`
		ProductName     string            `json:"productName"`
		ProductSum      int               `json:"ProductSum"`
	}

	var respOrders []OrderResponse
	db := global.DB
	var orders []models.Order
	db.Where("vz_store_id = ?", req.VzStoreID).Preload("Product").Preload("Expresses").Find(&orders)

	for _, order := range orders {
		newOrder := OrderResponse{
			VzOrderID:   order.VzOrderID,
			ProductID:   order.VzProductID,
			ProductName: order.Product.ProductName,
			ProductSum:  order.Quantity * order.Unit,
		}

		for _, exp := range order.Expresses {
			if req.Accepted != "" {
				// 如果Accepted不是空字符串验证筛选
				value, err := strconv.ParseInt(req.Accepted, 10, 8)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Accepted参数验证失败",
					})
					return
				}
				if exp.IsAccepted == int8(value) {
					logisticsResponse, err := services.QueryDelivery(exp.ExpressNumber)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": err.Error(),
						})
						return
					}
					newExp := ExpressResponse{
						ExpressNumber: exp.ExpressNumber,
						IsAccepted:    exp.IsAccepted,
						DeliveryTime:  exp.CreatedTime,
						Trace:         logisticsResponse.DataObj.LogisticsInfo.Traces,
					}
					newOrder.ExpressResponse = append(newOrder.ExpressResponse, newExp)
				}
			} else {
				// Accepted为空字符串
				logisticsResponse, err := services.QueryDelivery(exp.ExpressNumber)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": err.Error(),
					})
					return
				}
				newExp := ExpressResponse{
					ExpressNumber: exp.ExpressNumber,
					DeliveryTime:  exp.CreatedTime,
					Trace:         logisticsResponse.DataObj.LogisticsInfo.Traces,
				}
				newOrder.ExpressResponse = append(newOrder.ExpressResponse, newExp)
			}

		}
		respOrders = append(respOrders, newOrder)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "按验收状态筛选成功",
		"data":    respOrders,
	})
}

func GetLogisticsByNo(c *gin.Context) {
	type trackingQueryRequest struct {
		TrackingNumbers string `json:"tracking_numbers" binding:"required"`
	}
	var req trackingQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}
	trackingNumbers := strings.Split(req.TrackingNumbers, ",")

	var respExpress []services.LogisticsInfo
	for _, trackingNum := range trackingNumbers {
		logisticsResponse, err := services.QueryDelivery(strings.TrimSpace(trackingNum))
		if err != nil {
			respExpress = append(respExpress, services.LogisticsInfo{LogisticCode: strings.TrimSpace(trackingNum)})
		} else {
			respExpress = append(respExpress, logisticsResponse.DataObj.LogisticsInfo)
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "快递查询成功",
		"data":    respExpress,
	})
}

func UploadDelivery(c *gin.Context) {
	db := global.DB

	// 1. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请选择文件: " + err.Error(),
		})
		return
	}

	// 2. 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "打开文件失败: " + err.Error(),
		})
		return
	}
	defer src.Close()

	// 3. 读取Excel
	rows, err := tools.ReadExcel(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "读取Excel失败: " + err.Error(),
		})
		return
	}

	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Excel文件至少需要包含标题行和数据行",
		})
		return
	}

	// 4. 创建映射关系
	headerMap := make(map[string]int)
	for i, cell := range rows[0] {
		headerMap[cell] = i
	}

	// 在创建orders数组之前，先验证必需列
	requiredColumns := []string{"发货单号", "商品名称", "收货人", "电话", "地址", "物流单号"}
	var missingColumns []string
	for _, col := range requiredColumns {
		if _, ok := headerMap[col]; !ok {
			missingColumns = append(missingColumns, col)
		}
	}

	if len(missingColumns) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Excel缺少必需列，请确保包含以下必需列：%v", requiredColumns),
			"missing": missingColumns,
		})
		return
	}

	// 5. 处理数据行
	var orders []models.DeliveryOrder
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// 创建订单对象
		order := models.DeliveryOrder{
			DeliveryOrderNo:  getCellValue(row, headerMap, "发货单号"),
			StoreID:          getCellValue(row, headerMap, "门店ID"),
			StoreName:        getCellValue(row, headerMap, "门店名称"),
			Period:           getCellValue(row, headerMap, "期数"),
			ProductID:        getCellValue(row, headerMap, "商品ID"),
			ProductName:      getCellValue(row, headerMap, "商品名称"),
			Specification:    getCellValue(row, headerMap, "规格"),
			Supplier:         getCellValue(row, headerMap, "供应商"),
			Status:           getCellValue(row, headerMap, "状态"),
			Receiver:         getCellValue(row, headerMap, "收货人"),
			Phone:            getCellValue(row, headerMap, "电话"),
			Address:          getCellValue(row, headerMap, "地址"),
			LogisticsCompany: getCellValue(row, headerMap, "物流公司"),
			TrackingNumber:   getCellValue(row, headerMap, "物流单号"),
		}

		// 处理数值字段
		if idx, ok := headerMap["实发单数"]; ok && idx < len(row) {
			order.DeliveryCount, _ = strconv.Atoi(row[idx])
		}
		if idx, ok := headerMap["实发商品数"]; ok && idx < len(row) {
			order.DeliveryQuantity, _ = strconv.Atoi(row[idx])
		}

		// 跳过空行
		if order.DeliveryOrderNo == "" {
			continue
		}

		orders = append(orders, order)
	}

	if len(orders) > 0 {
		// 开始事务
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 在事务中使用 CreateInBatches
		batchSize := 500
		if err := tx.CreateInBatches(&orders, batchSize).Error; err != nil {
			tx.Rollback()

			// 原有的错误处理逻辑...
			var duplicateOrderNo string
			if strings.Contains(err.Error(), "Duplicate entry") {
				re := regexp.MustCompile(`Duplicate entry '([^']+)' for key`)
				matches := re.FindStringSubmatch(err.Error())
				if len(matches) > 1 {
					duplicateOrderNo = matches[1]
				}
			}

			errorMessage := "保存数据失败"
			if duplicateOrderNo != "" {
				errorMessage = fmt.Sprintf("发货单号 '%s' 已存在，请勿重复上传。", duplicateOrderNo)
			}

			log.Printf("数据库保存失败: %v", err)

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errorMessage,
				"detail":  "上传的文件中包含已存在的记录，所有数据均未保存",
			})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			log.Printf("事务提交失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "数据保存失败，请重试",
			})
			return
		}
	}

	// 7. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"message": "上传成功",
		"count":   len(orders),
	})
}

// 辅助函数：获取单元格值
func getCellValue(row []string, headerMap map[string]int, key string) string {
	if idx, ok := headerMap[key]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func GetLogisticsByPhone(c *gin.Context) {
	// 获取查询参数
	phone := c.Query("phone")

	// 如果phone为空，显示搜索页面
	if phone == "" {
		c.HTML(http.StatusOK, "fahuo.html", gin.H{
			"success": true,
			"phone":   "",
			"orders":  []models.DeliveryOrder{},
		})
		return
	}

	// 查询数据库，按照创建时间降序排列
	var orders []models.DeliveryOrder
	if err := global.DB.Where("phone = ?", phone).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.HTML(http.StatusOK, "fahuo.html", gin.H{
			"success": false,
			"error":   "查询失败: " + err.Error(),
			"phone":   phone,
			"orders":  []models.DeliveryOrder{},
		})
		return
	}

	// 返回结果
	c.HTML(http.StatusOK, "fahuo.html", gin.H{
		"success": true,
		"phone":   phone,
		"orders":  orders,
		"count":   len(orders),
	})
}
