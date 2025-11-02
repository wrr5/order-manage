package tools

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte("dating-secret-key")

// 生成 JWT Token
func GenerateJWT(userID uint, mobile string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(72 * time.Hour) // 72小时过期

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"mobile":  mobile,
		"exp":     expireTime.Unix(),
	})
	return token.SignedString(jwtSecret)
}

// 解析JWT Token
func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 token 是否有效
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("无效的token")
	}
}

type Pagination struct {
	CurrentPage     int
	TotalPages      int
	PageSize        int
	BasePath        string
	DisplayPages    []int // 要显示的页码列表
	HasPrevEllipsis bool  // 前面是否有省略号
	HasNextEllipsis bool  // 后面是否有省略号
}

// 在结构体上定义方法
func (p Pagination) NextPage() int {
	if p.CurrentPage >= p.TotalPages {
		return p.TotalPages
	}
	return p.CurrentPage + 1
}

func (p Pagination) PrevPage() int {
	if p.CurrentPage <= 1 {
		return 1
	}
	return p.CurrentPage - 1
}

func (p *Pagination) CalculateDisplayPages(maxDisplay int) {
	if p.TotalPages <= maxDisplay {
		// 总页数小于等于最大显示数量，显示所有页码
		p.DisplayPages = make([]int, p.TotalPages)
		for i := 0; i < p.TotalPages; i++ {
			p.DisplayPages[i] = i + 1
		}
		return
	}

	// 计算起始和结束页码
	start := p.CurrentPage - maxDisplay/2
	end := p.CurrentPage + maxDisplay/2

	if start < 1 {
		start = 1
		end = maxDisplay
	}

	if end > p.TotalPages {
		end = p.TotalPages
		start = end - maxDisplay + 1
	}

	// 生成显示的页码列表
	p.DisplayPages = make([]int, end-start+1)
	for i := range p.DisplayPages {
		p.DisplayPages[i] = start + i
	}

	// 判断是否需要省略号
	p.HasPrevEllipsis = start > 1
	p.HasNextEllipsis = end < p.TotalPages
}
