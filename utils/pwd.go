package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 接收明文密码，并返回哈希后的密码
func HashPassword(password string) (string, error) {
	// 使用 bcrypt 加盐并哈希密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 验证明文密码与哈希密码是否匹配
func CheckPassword(password, hashedPassword string) bool {
	// 使用 bcrypt 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
