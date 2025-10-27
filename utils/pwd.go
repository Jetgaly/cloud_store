package utils

import (
    "golang.org/x/crypto/bcrypt"
)

// HashPassword 安全地哈希密码
// 会自动生成随机盐，并将盐和哈希组合在一起返回
func HashPassword(password string) (string, error) {
    // cost 是工作因子，越高越安全但也越慢(4-31)
    // 推荐值：bcrypt.DefaultCost 或更高
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

// CheckPasswordHash 验证密码是否匹配
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}