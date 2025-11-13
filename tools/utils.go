package tools

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte("general-management-secret-key")

// 完整的JWT工具函数
type JWTClaims struct {
	UserID int
	Phone  string
	jwt.StandardClaims
}

func GenerateJWT(userID int, phone string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(72 * time.Hour)

	claims := JWTClaims{
		UserID: userID,
		Phone:  phone,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  nowTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("无效的token")
	}
}
