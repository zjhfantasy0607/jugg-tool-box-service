package middleware

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取authorization header
		tokenString := c.GetHeader("Authorization")

		// validate token formate
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 4010, "msg": "权限不足"})
			c.Abort()
			return
		}

		tokenString = tokenString[7:]
		token, claims, err := common.ParseToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 4011, "msg": "权限不足"})
			c.Abort()
			return
		}

		// 验证通过后获取 claim 中的 userId
		uid := claims.UID
		DB := common.GetDB()
		var user model.User
		DB.Where("uid = ?", uid).First(&user)
		// 用户不存在
		if user.ID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 4012, "msg": "权限不足"})
			c.Abort()
			return
		}

		// 用户存在 将 user 的信息写入上下文
		c.Set("user", user)
		c.Next()
	}
}
