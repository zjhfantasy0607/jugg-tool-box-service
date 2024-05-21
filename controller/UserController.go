package controller

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	DB := common.GetDB()

	/***** 获取参数 *****/
	name := c.PostForm("name")
	telephone := c.PostForm("telephone")
	password := c.PostForm("password")

	/***** 数据验证 *****/
	if len(telephone) != 11 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "手机号必须为11位"})
		return
	}
	if len(password) < 6 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "密码不能少于6位"})
		return
	}
	if len(name) == 0 {
		name = util.RandomString(10)
	}

	/***** 判断手机号是否存在 *****/
	if isTelephoneExist(DB, telephone) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "用户已存在"})
		return
	}

	/***** 创建用户 *****/
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "加密错误"})
		return
	}

	DB.Create(&model.User{
		Name:      name,
		Telephone: telephone,
		Password:  string(hashedPassword),
	})

	// 返回结果
	c.JSON(200, gin.H{
		"code":    200,
		"message": "注册成功",
	})
}

func Login(c *gin.Context) {
	DB := common.GetDB()

	/***** 获取用户 *****/
	telephone := c.PostForm("telephone")
	password := c.PostForm("password")

	/***** 数据验证 *****/
	if len(telephone) != 11 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "手机号必须为11位"})
		return
	}
	if len(password) < 6 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "密码不能少于6位"})
		return
	}

	/***** 判断手机号是否存在 *****/
	var user model.User
	DB.Where("telephone = ?", telephone).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "用户不存在"})
		return
	}

	/***** 判断密码是否正确 *****/
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 400, "msg": "密码错误"})
		return
	}

	/***** 发放token *****/
	token := 11

	c.JSON(200, gin.H{"code": 200, "msg": "登录成功", "data": gin.H{"token": token}})
}

func isTelephoneExist(db *gorm.DB, telephone string) bool {
	var user model.User
	db.Where("telephone = ?", telephone).First(&user)

	return user.ID != 0
}
