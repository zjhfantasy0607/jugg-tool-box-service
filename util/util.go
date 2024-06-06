package util

import (
	"math/rand"
	"regexp"
	"time"
)

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func IsEmail(email string) bool {
	// 正则表达式用来匹配邮箱地址的模式
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// 编译正则表达式
	regex := regexp.MustCompile(pattern)
	// 判断是否匹配
	return regex.MatchString(email)
}
