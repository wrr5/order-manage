package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wrr5/order-manage/global"
	"github.com/wrr5/order-manage/models"
	"github.com/wrr5/order-manage/services"
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

func GetProduct(c *gin.Context) {
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

func GetProductById(c *gin.Context) {
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
