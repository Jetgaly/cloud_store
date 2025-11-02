package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义自定义 Claims，可以包含标准 Claims 和自定义字段
type CustomClaims struct {
	UserId   int    `json:"user_id"`
	NickName string `json:"nickname"`
	jwt.RegisteredClaims
}

// JWT 配置
var (
	jwtSecret = []byte("SecretKey")
)

// GenerateToken 生成 JWT Token
func GenerateToken(userID int, NickName string) (string, error) {
	// 设置 token 过期时间
	expirationTime := time.Now().Add(30 * time.Minute)

	// 创建 Claims
	claims := CustomClaims{
		UserId:   userID,
		NickName: NickName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			//IssuedAt:  jwt.NewNumericDate(time.Now()),     // 签发时间
			//NotBefore: jwt.NewNumericDate(time.Now()),     // 生效时间
			Issuer:  "CloudStore", // 签发者
			Subject: "user-auth",  // 主题
			//ID:        fmt.Sprintf("%d", userID),          // token ID
		},
	}

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 生成签名字符串
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("生成 token 失败: %s", err.Error())
	}

	return tokenString, nil
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析 token
	var claims CustomClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析 token 失败: %v", err)
	}

	// 验证 token 是否有效
	if token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("无效的 token")
}
