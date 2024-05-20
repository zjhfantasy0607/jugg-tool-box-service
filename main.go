package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/api/auth/register", func(c *gin.Context) {
		// 获取参数
		name := c.PostForm("name")
		telephone := c.PostForm("telephone")
		password := c.PostForm("password")

		// 数据验证
		if len(telephone) != 11 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "手机号必须为11位"})
			return
		}
		if len(password) < 6 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "密码不能少于6位"})
			return
		}
		if len(name) == 0 {
			name = RandomString(10)
		}

		log.Println(name, telephone, password)

		// 判断手机号是否存在

		// 创建用户

		// 返回结果
		c.JSON(200, gin.H{
			"message": "注册成功",
		})
	})

	panic(r.Run()) // listen and serve on 0.0.0.0:8080
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
