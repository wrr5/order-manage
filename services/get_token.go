package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 定义响应结构体
type TokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Authorization string `json:"authorization"`
		ZbId          string `json:"zbId"`
		Liveticket    string `json:"liveticket"`
	} `json:"data"`
}

func GetToken() string {
	// 请求URL
	url := "https://www.hzsh.online/api/get_url/1"

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("创建请求错误: %v\n", err)
		return ""
	}

	// 设置请求头
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Mobile Safari/537.36")

	// 创建HTTP客户端并发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求错误: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		return ""
	}

	// 解析JSON响应
	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		fmt.Printf("解析JSON错误: %v\n", err)
		return ""
	}

	// 检查响应码
	if tokenResponse.Code != 200 {
		fmt.Printf("API返回错误: %s\n", tokenResponse.Message)
		return ""
	}

	// 提取authorization
	return tokenResponse.Data.Authorization
}
