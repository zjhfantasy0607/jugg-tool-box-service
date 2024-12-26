package util

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func RandomNumberString(length int) string {
	const charset = "0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func Md5Hash(data []byte) (string, error) {
	hash := md5.New()
	_, err := hash.Write(data)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func SplitLevelPath(path string) []string {
	// 分割字符串为切片
	parts := strings.Split(path, "/")
	var result []string
	var currentPath string

	// 遍历分割后的切片，并逐步拼接路径
	for _, part := range parts {
		if part == "" {
			// 跳过空字符串部分 (因为 Split 会把前导 "/" 产生的空字符串保留)
			continue
		}
		if currentPath == "" {
			// 第一个部分
			currentPath = "/" + part
		} else {
			// 拼接后续路径
			currentPath = currentPath + "/" + part
		}
		// 将拼接后的路径添加到结果切片中
		result = append(result, currentPath)
	}

	return result
}
