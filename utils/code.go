package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

// GenerateOrderCode 生成订单编号（格式：YYYYMMDD-随机码）
func GenerateOrderCode() string {
	// 获取当前日期（YYYYMMDD）
	datePart := time.Now().Format("20060102")

	// 生成 4 字节随机数（提高唯一性）
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)

	// Base32 编码（去掉填充 = 号）
	randomPart := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	randomPart = strings.TrimRight(randomPart, "=")[:6] // 截取前 6 个字符

	// 组合日期 + 随机部分
	return fmt.Sprintf("%s-%s", datePart, randomPart)
}
