package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func CalculateSHA256Stream(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	// 使用 io.Copy 自动处理缓冲区
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("计算哈希失败: %v", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
