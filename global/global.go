package global

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type TokenManager struct {
	token      string
	lastUpdate time.Time
	lock       sync.RWMutex
}

var (
	DB *gorm.DB
	TM = &TokenManager{}
)

// refresh 刷新token
func (tm *TokenManager) refresh() {
	token := GetToken() // 调用你的GetToken函数

	tm.lock.Lock()
	defer tm.lock.Unlock()

	tm.token = token
	tm.lastUpdate = time.Now()

	log.Printf("token刷新成功, 更新时间: %s", tm.lastUpdate.Format("2006-01-02 15:04:05"))
}

// Get 获取当前token
func (tm *TokenManager) Get() string {
	tm.lock.RLock()
	defer tm.lock.RUnlock()
	return tm.token
}

// ShouldRefresh 检查是否需要刷新
func (tm *TokenManager) ShouldRefresh() bool {
	tm.lock.RLock()
	defer tm.lock.RUnlock()
	return tm.token == "" || time.Since(tm.lastUpdate) > 8*time.Hour
}

// StartAutoRefresh 启动自动刷新（包含初始化）
func (tm *TokenManager) StartAutoRefresh() {
	// 启动时立即刷新一次（初始化）
	tm.refresh()

	// 每8小时刷新一次
	go func() {
		ticker := time.NewTicker(8 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("定时刷新token...")
			tm.refresh()
		}
	}()
}

// ForceRefresh 强制刷新token
func (tm *TokenManager) ForceRefresh() {
	log.Println("强制刷新token...")
	tm.refresh()
}
