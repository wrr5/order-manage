package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wrr5/order-manage/global"
)

type LogisticsResponse struct {
	IsOK    bool    `json:"isok"`
	Msg     string  `json:"msg"`
	Code    int     `json:"code"`
	DataObj DataObj `json:"dataObj"`
	Amout   int     `json:"amout"`
}

type DataObj struct {
	LogisticsInfo   LogisticsInfo   `json:"logisticsInfo"`
	DestinationInfo DestinationInfo `json:"destinationInfo"`
	ShipperInfo     ShipperInfo     `json:"shipperInfo"`
	Remark          string          `json:"remark"`
}

type LogisticsInfo struct {
	ShipperCode  string  `json:"shipperCode"`  // 快递公司编码
	LogisticCode string  `json:"logisticCode"` // 物流单号
	State        string  `json:"state"`        // 状态码
	StateText    string  `json:"stateText"`    // 状态文本
	StateEx      string  `json:"stateEx"`      // 扩展状态码
	StateExText  string  `json:"stateExText"`  // 扩展状态文本
	Location     string  `json:"location"`     // 当前位置
	Traces       []Trace `json:"traces"`       // 物流轨迹
}

type Trace struct {
	AcceptStation string `json:"acceptStation"` // 物流描述
	AcceptTime    string `json:"acceptTime"`    // 时间
	Action        string `json:"action"`        // 动作码
	ActionText    string `json:"actionText"`    // 动作文本
	StateText     string `json:"stateText"`     // 状态文本
}

type DestinationInfo struct {
	ReceiveUserAddress string `json:"receiveUserAddress"` // 收件人地址
	ReceiveUserName    string `json:"receiveUserName"`    // 收件人姓名
	ReceiveUserPhone   string `json:"receiveUserPhone"`   // 收件人电话
}

type ShipperInfo struct {
	DeliveryName string      `json:"deliveryName"` // 快递公司名称
	DeliveryNo   string      `json:"deliveryNo"`   // 快递单号
	DeliveryTime interface{} `json:"deliveryTime"` // 发货时间（可能是 null）
	HasDelivery  bool        `json:"hasDelivery"`  // 是否有物流信息
}

// 壹米滴答物流特殊处理
func ifyimidida(s string) bool {
	return len(s) == 12 && strings.HasPrefix(s, "112")
}

func QueryDelivery(expressNumber string) (LogisticsResponse, error) {
	var queryUrl string
	if ifyimidida(expressNumber) {
		// 对中文参数值进行URL编码
		deliveryName := "壹米滴答"
		encodedDeliveryName := url.QueryEscape(deliveryName)
		queryUrl = fmt.Sprintf("https://shop.vzan.com/api/zbdeliveryprocure/getlogistics?storeId=123456&deliveryNo=%s&deliveryName=%s", expressNumber, encodedDeliveryName)
	} else {
		queryUrl = fmt.Sprintf("https://shop.vzan.com/api/zbdeliveryprocure/getlogistics?storeId=123456&deliveryNo=%s", expressNumber)
	}
	req, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		return LogisticsResponse{}, fmt.Errorf("创建请求错误: %v", err)
	}
	// 设置请求头
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", global.TM.Get())
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Mobile Safari/537.36")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return LogisticsResponse{}, fmt.Errorf("请求错误: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LogisticsResponse{}, fmt.Errorf("读取响应失败: %v", err)
	}
	var logisticsResponse LogisticsResponse

	if err := json.Unmarshal(body, &logisticsResponse); err != nil {
		return LogisticsResponse{}, fmt.Errorf("解析响应失败: %v", err)
	}
	if len(logisticsResponse.DataObj.LogisticsInfo.Traces) < 1 {
		return LogisticsResponse{}, fmt.Errorf("%s暂无物流信息", expressNumber)
	}

	return logisticsResponse, nil
}
