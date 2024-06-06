package controller

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/dto"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/util"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Regist(c *gin.Context) {
	DB := common.GetDB()

	/***** 获取参数 *****/
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")

	/***** 数据验证 *****/
	if !util.IsEmail(email) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "邮箱格式错误")
		return
	}
	if len(password) < 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "密码不能少于6位")
		return
	}
	if len(name) == 0 {
		name = util.RandomString(10)
	}

	/***** 判断手机号是否存在 *****/
	if isTelephoneExist(DB, email) {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "用户已存在")
		return
	}

	/***** 创建用户 *****/
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "加密错误")
		return
	}

	DB.Create(&model.User{
		Name:      name,
		Telephone: email,
		Password:  string(hashedPassword),
	})

	// 返回结果
	response.Success(c, nil, "注册成功")
}

func SignIn(c *gin.Context) {
	DB := common.GetDB()

	/***** 获取用户 *****/
	email := c.PostForm("email")
	password := c.PostForm("password")

	/***** 数据验证 *****/
	if !util.IsEmail(email) {
		response.Response(c, http.StatusUnprocessableEntity, 422, gin.H{"email": email}, "邮箱格式错误")
		return
	}
	if len(password) < 6 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "密码不能少于6位")
		return
	}

	/***** 判断邮箱是否存在 *****/
	var user model.User
	DB.Where("email = ?", email).First(&user)
	if user.ID == 0 {
		response.Response(c, http.StatusUnprocessableEntity, 422, nil, "用户不存在")
		return
	}

	/***** 判断密码是否正确 *****/
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		response.Response(c, http.StatusUnprocessableEntity, 400, nil, "密码错误")
		return
	}

	/***** 发放token *****/
	token, err := common.ReleaseToken(user)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error : %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "登录成功")
}

func Info(c *gin.Context) {
	user, _ := c.Get("user")
	response.Success(c, gin.H{"user": dto.ToUserDto(user.(model.User))}, "")
}

func isTelephoneExist(db *gorm.DB, telephone string) bool {
	var user model.User
	db.Where("telephone = ?", telephone).First(&user)

	return user.ID != 0
}
