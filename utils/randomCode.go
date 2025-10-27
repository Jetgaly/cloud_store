package utils

import (
	"fmt"
	"math/rand"
)

func GenerateCode() string {
	code := rand.Intn(100000)        // 生成 0-99999 的随机数
	return fmt.Sprintf("%05d", code) // 格式化为 5 位，不足补零
}