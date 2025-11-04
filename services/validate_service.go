package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ApiResponse struct {
	IsOK      bool   `json:"isok"`
	Msg       string `json:"Msg"`
	Code      int    `json:"code"`
	DataObj   Data   `json:"dataObj"`
	RequestID string `json:"requestId"`
	Message   string `json:"msg"`
}

type Data struct {
	Records          []Record `json:"records"`
	Total            int      `json:"total"`
	Size             int      `json:"size"`
	Current          int      `json:"current"`
	Orders           []string `json:"orders"` // 根据实际数据结构可能需要调整
	OptimizeCountSql bool     `json:"optimizeCountSql"`
	HitCount         bool     `json:"hitCount"`
	CountID          *string  `json:"countId"`
	MaxLimit         *int     `json:"maxLimit"`
	SearchCount      bool     `json:"searchCount"`
	Pages            int      `json:"pages"`
}

type Record struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"userId"`
	Nickname         string     `json:"nickname"`
	HeadImgURL       string     `json:"headImgUrl"`
	Name             string     `json:"name"`
	Phone            string     `json:"phone"`
	AreaName         string     `json:"areaName"`
	AuditTime        string     `json:"auditTime"`
	Remark           string     `json:"remark"`
	State            int        `json:"state"`
	TopicNum         int        `json:"topicNum"`
	AreaCode         string     `json:"areaCode"`
	InviteCode       string     `json:"inviteCode"`
	PopuTopicNum     int        `json:"popuTopicNum"`
	InviteNum        int        `json:"inviteNum"`
	PopuPeopleNum    int        `json:"popuPeopleNum"`
	RedPackageAmount int        `json:"redPackageAmount"`
	Level            int        `json:"level"`
	LevelName        string     `json:"levelName"`
	ParentID         int64      `json:"parentId"`
	ParentUserID     int64      `json:"parentUserId"`
	ParentName       string     `json:"parentName"`
	ParentPhone      string     `json:"parentPhone"`
	SupName          string     `json:"supName"`
	SupPhone         string     `json:"supPhone"`
	GeneralUserID    int64      `json:"generalUserId"`
	CustomerNum      int        `json:"customerNum"`
	CategoryID       int64      `json:"categoryId"`
	CategoryName     string     `json:"categoryName"`
	CorpID           *string    `json:"corpId"`
	WeComAccount     *string    `json:"weComAccount"`
	StoreID          int64      `json:"storeId"`
	StoreName        *string    `json:"storeName"`
	GroupID          *int       `json:"groupId"`
	GroupName        *string    `json:"groupName"`
	WeWorkUserInfo   *string    `json:"weWorkUserInfo"`
	Country          string     `json:"country"`
	AgentPeriods     *string    `json:"agentPeriods"`
	TeamName         *string    `json:"teamName"`
	IsTeamLeader     *bool      `json:"isTeamLeader"`
	LeaderTeamName   *string    `json:"leaderTeamName"`
	CustomInfo       CustomInfo `json:"customInfo"`
	StoreAddress     *string    `json:"storeAddress"`
	GroupChatState   int        `json:"groupChatState"`
	AccountState     int        `json:"accountState"`
	Source           int        `json:"source"`
}

type CustomInfo struct {
	Level   int      `json:"level"`
	List    []string `json:"list"` // 根据实际内容可能需要调整类型
	ImgList *string  `json:"imgList"`
}

// 验证注册信息是否有效
func ValidatePhoneName(phone, name string) (bool, ApiResponse, error) {
	// 请求url
	url := "https://live-gw.vzan.com/health/v1/admin/agent/pageAgent"
	// 构建请求体
	requestBody := map[string]interface{}{
		"phone": phone,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return false, ApiResponse{}, fmt.Errorf("JSON编码错误: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, ApiResponse{}, fmt.Errorf("创建请求错误: %v", err)
	}

	// 设置请求头
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", GetToken())
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Mobile Safari/537.36")

	// 创建HTTP客户端并发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, ApiResponse{}, fmt.Errorf("请求错误: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, ApiResponse{}, fmt.Errorf("读取响应失败: %v", err)
	}
	// 解析响应
	var apiResponse ApiResponse

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return false, ApiResponse{}, fmt.Errorf("解析响应失败: %v", err)
	}

	// 根据第三方API的响应逻辑判断注册信息是否有效
	if !apiResponse.IsOK {
		errorMsg := "未知错误"
		if apiResponse.Msg != "" {
			errorMsg = apiResponse.Msg
		}
		return false, ApiResponse{}, fmt.Errorf("API请求失败: %s (错误码: %d)", errorMsg, apiResponse.Code)
	}

	if apiResponse.DataObj.Total < 1 {
		return false, ApiResponse{}, fmt.Errorf("无效的手机号, 响应数据中未找到记录")
	}

	if apiResponse.DataObj.Records[0].Name != name {
		return false, ApiResponse{}, fmt.Errorf("姓名与手机号不匹配")
	}

	return true, apiResponse, nil
}
